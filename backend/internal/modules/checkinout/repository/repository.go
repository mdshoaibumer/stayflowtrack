package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stayflow/stayflow-track/internal/modules/checkinout/domain"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// ReservationInfo holds info needed for check-in/out operations.
type ReservationInfo struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	PropertyID   uuid.UUID
	UnitID       uuid.UUID
	GuestID      uuid.UUID
	Status       string
	CheckInDate  time.Time
	CheckOutDate time.Time
	RatePerNight decimal.Decimal
	TotalAmount  decimal.Decimal
	GuestName    string
	UnitNumber   string
}

func (r *Repository) GetReservationForCheckIn(ctx context.Context, reservationID, tenantID uuid.UUID) (*ReservationInfo, error) {
	var info ReservationInfo
	err := r.pool.QueryRow(ctx,
		`SELECT r.id, r.tenant_id, r.property_id, r.unit_id, r.guest_id,
		        r.status, r.check_in_date, r.check_out_date, r.rate_per_night, r.total_amount,
		        g.first_name || ' ' || g.last_name, u.unit_number
		 FROM reservations r
		 JOIN guests g ON r.guest_id = g.id
		 JOIN units u ON r.unit_id = u.id
		 WHERE r.id = $1 AND r.tenant_id = $2`,
		reservationID, tenantID,
	).Scan(&info.ID, &info.TenantID, &info.PropertyID, &info.UnitID, &info.GuestID,
		&info.Status, &info.CheckInDate, &info.CheckOutDate, &info.RatePerNight, &info.TotalAmount,
		&info.GuestName, &info.UnitNumber)

	if err == pgx.ErrNoRows {
		return nil, apperrors.NotFound("reservation", reservationID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("get reservation: %w", err)
	}
	return &info, nil
}

// PerformCheckIn executes check-in atomically: update reservation, unit status, create folio, record deposit.
func (r *Repository) PerformCheckIn(ctx context.Context, tenantID uuid.UUID, input domain.CheckInInput, info *ReservationInfo, userID uuid.UUID) (*domain.CheckInDetails, uuid.UUID, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 1. Update reservation status to checked_in
	result, err := tx.Exec(ctx,
		`UPDATE reservations SET status = 'checked_in', actual_check_in = NOW(), unit_id = $3
		 WHERE id = $1 AND tenant_id = $2 AND status = 'confirmed'`,
		input.ReservationID, tenantID, input.AssignedUnitID,
	)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("update reservation: %w", err)
	}
	if result.RowsAffected() == 0 {
		return nil, uuid.Nil, apperrors.BadRequest("reservation must be in 'confirmed' status to check in")
	}

	// 2. Update unit status to occupied
	_, err = tx.Exec(ctx,
		`UPDATE units SET status = 'occupied' WHERE id = $1 AND tenant_id = $2`,
		input.AssignedUnitID, tenantID,
	)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("update unit status: %w", err)
	}

	// If assigned unit differs from original, release original
	if input.AssignedUnitID != info.UnitID {
		_, err = tx.Exec(ctx,
			`UPDATE units SET status = 'available' WHERE id = $1 AND tenant_id = $2`,
			info.UnitID, tenantID,
		)
		if err != nil {
			return nil, uuid.Nil, fmt.Errorf("release original unit: %w", err)
		}
	}

	// 3. Create folio
	var folioID uuid.UUID
	var folioNumber string
	err = tx.QueryRow(ctx,
		`INSERT INTO folios (tenant_id, reservation_id, guest_id, property_id, folio_number, status)
		 VALUES ($1, $2, $3, $4, generate_folio_number($1), 'open')
		 RETURNING id, folio_number`,
		tenantID, input.ReservationID, info.GuestID, info.PropertyID,
	).Scan(&folioID, &folioNumber)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("create folio: %w", err)
	}

	// 4. Add room charges to folio
	nights := int(info.CheckOutDate.Sub(info.CheckInDate).Hours() / 24)
	roomAmount := info.RatePerNight.Mul(decimal.NewFromInt(int64(nights)))
	taxRate := decimal.NewFromInt(18) // Default GST rate for hotels > ₹7500
	if info.RatePerNight.LessThan(decimal.NewFromInt(7500)) {
		taxRate = decimal.NewFromInt(12)
	}
	if info.RatePerNight.LessThan(decimal.NewFromInt(1000)) {
		taxRate = decimal.Zero
	}
	taxAmount := roomAmount.Mul(taxRate).Div(decimal.NewFromInt(100)).Round(2)
	roomTotal := roomAmount.Add(taxAmount)

	_, err = tx.Exec(ctx,
		`INSERT INTO line_items (tenant_id, folio_id, category, description, quantity, unit_price, amount, tax_rate, tax_amount, total, date, created_by)
		 VALUES ($1, $2, 'room_charge', $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		tenantID, folioID,
		fmt.Sprintf("Room charges - %d night(s)", nights),
		nights, info.RatePerNight, roomAmount, taxRate, taxAmount, roomTotal,
		info.CheckInDate, userID,
	)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("add room charge: %w", err)
	}

	// 5. Record deposit if provided
	if input.DepositAmount.IsPositive() {
		_, err = tx.Exec(ctx,
			`INSERT INTO payments (tenant_id, folio_id, payment_type, payment_method, amount, reference_number, received_by)
			 VALUES ($1, $2, 'deposit', $3, $4, $5, $6)`,
			tenantID, folioID, input.DepositMethod, input.DepositAmount, input.DepositReference, userID,
		)
		if err != nil {
			return nil, uuid.Nil, fmt.Errorf("record deposit: %w", err)
		}
	}

	// 6. Update folio totals
	_, err = tx.Exec(ctx,
		`UPDATE folios SET
			subtotal = $2, tax_total = $3, total_amount = $4,
			paid_amount = $5, balance = $6
		 WHERE id = $1`,
		folioID, roomAmount, taxAmount, roomTotal, input.DepositAmount, roomTotal.Sub(input.DepositAmount),
	)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("update folio totals: %w", err)
	}

	// 7. Record check-in details
	var details domain.CheckInDetails
	err = tx.QueryRow(ctx,
		`INSERT INTO check_in_details (tenant_id, reservation_id, assigned_unit_id, deposit_amount, deposit_method, deposit_reference, id_document_type, checked_in_by, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, created_at`,
		tenantID, input.ReservationID, input.AssignedUnitID,
		input.DepositAmount, input.DepositMethod, input.DepositReference,
		input.IDDocumentType, userID, input.Notes,
	).Scan(&details.ID, &details.CreatedAt)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("record check-in details: %w", err)
	}

	details.TenantID = tenantID
	details.ReservationID = input.ReservationID
	details.AssignedUnitID = input.AssignedUnitID
	details.DepositAmount = input.DepositAmount
	details.DepositMethod = input.DepositMethod
	details.DepositReference = input.DepositReference
	details.IDDocumentType = input.IDDocumentType
	details.CheckedInBy = userID
	details.Notes = input.Notes

	if err := tx.Commit(ctx); err != nil {
		return nil, uuid.Nil, fmt.Errorf("commit: %w", err)
	}

	return &details, folioID, nil
}

// PerformCheckOut executes check-out atomically: close folio, generate invoice, release unit.
func (r *Repository) PerformCheckOut(ctx context.Context, tenantID uuid.UUID, info *ReservationInfo, userID uuid.UUID, notes string) (*domain.CheckOutResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 1. Update reservation status
	result, err := tx.Exec(ctx,
		`UPDATE reservations SET status = 'checked_out', actual_check_out = NOW()
		 WHERE id = $1 AND tenant_id = $2 AND status = 'checked_in'`,
		info.ID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("update reservation: %w", err)
	}
	if result.RowsAffected() == 0 {
		return nil, apperrors.BadRequest("reservation must be in 'checked_in' status to check out")
	}

	// 2. Release unit (set to cleaning)
	_, err = tx.Exec(ctx,
		`UPDATE units SET status = 'cleaning' WHERE id = $1 AND tenant_id = $2`,
		info.UnitID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("release unit: %w", err)
	}

	// 3. Get folio
	var folioID uuid.UUID
	var subtotal, taxTotal, totalAmount, paidAmount, balance decimal.Decimal
	err = tx.QueryRow(ctx,
		`SELECT id, subtotal, tax_total, total_amount, paid_amount, balance
		 FROM folios WHERE reservation_id = $1 AND tenant_id = $2 AND status = 'open'`,
		info.ID, tenantID,
	).Scan(&folioID, &subtotal, &taxTotal, &totalAmount, &paidAmount, &balance)
	if err == pgx.ErrNoRows {
		return nil, apperrors.BadRequest("no open folio found for this reservation")
	}
	if err != nil {
		return nil, fmt.Errorf("get folio: %w", err)
	}

	// 4. Close folio
	_, err = tx.Exec(ctx,
		`UPDATE folios SET status = 'closed', closed_at = NOW() WHERE id = $1`,
		folioID,
	)
	if err != nil {
		return nil, fmt.Errorf("close folio: %w", err)
	}

	// 5. Generate invoice
	nights := int(info.CheckOutDate.Sub(info.CheckInDate).Hours() / 24)
	cgst := taxTotal.Div(decimal.NewFromInt(2)).Round(2)
	sgst := taxTotal.Div(decimal.NewFromInt(2)).Round(2)

	var invoiceID uuid.UUID
	err = tx.QueryRow(ctx,
		`INSERT INTO invoices (
			tenant_id, folio_id, reservation_id, guest_id, property_id,
			invoice_number, status, guest_name, guest_phone, property_name, property_address,
			subtotal, cgst_amount, sgst_amount, total_tax, total_amount,
			paid_amount, balance_due, check_in_date, check_out_date, num_nights
		) VALUES (
			$1, $2, $3, $4, $5,
			generate_invoice_number($1), 'issued', $6, $7, $8, $9,
			$10, $11, $12, $13, $14,
			$15, $16, $17, $18, $19
		) RETURNING id`,
		tenantID, folioID, info.ID, info.GuestID, info.PropertyID,
		info.GuestName, "", "", "", // phone/property details filled from query below
		subtotal, cgst, sgst, taxTotal, totalAmount,
		paidAmount, balance, info.CheckInDate, info.CheckOutDate, nights,
	).Scan(&invoiceID)
	if err != nil {
		return nil, fmt.Errorf("create invoice: %w", err)
	}

	// 6. Copy line items to invoice
	_, err = tx.Exec(ctx,
		`INSERT INTO invoice_line_items (invoice_id, category, description, quantity, unit_price, amount, tax_rate, tax_amount, total, date)
		 SELECT $1, category, description, quantity, unit_price, amount, tax_rate, tax_amount, total, date
		 FROM line_items WHERE folio_id = $2 AND is_void = false`,
		invoiceID, folioID,
	)
	if err != nil {
		return nil, fmt.Errorf("copy line items: %w", err)
	}

	// 7. Increment guest stays
	_, err = tx.Exec(ctx, `UPDATE guests SET total_stays = total_stays + 1 WHERE id = $1`, info.GuestID)
	if err != nil {
		return nil, fmt.Errorf("increment guest stays: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &domain.CheckOutResult{
		ReservationID:  info.ID,
		GuestName:      info.GuestName,
		UnitNumber:     info.UnitNumber,
		CheckInDate:    info.CheckInDate,
		CheckOutDate:   info.CheckOutDate,
		ActualCheckOut: time.Now(),
		TotalCharges:   totalAmount,
		TotalPayments:  paidAmount,
		Balance:        balance,
		InvoiceID:      invoiceID,
	}, nil
}
