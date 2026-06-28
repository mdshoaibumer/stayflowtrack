package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"github.com/stayflow/stayflow-track/internal/modules/billing/domain"
	"github.com/stayflow/stayflow-track/internal/platform/database"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

type Repository struct {
	pool *database.TenantPool
}

func New(pool *database.TenantPool) *Repository {
	return &Repository{pool: pool}
}

// GetFolioByID returns a folio with tenant isolation.
func (r *Repository) GetFolioByID(ctx context.Context, folioID, tenantID uuid.UUID) (*domain.Folio, error) {
	var f domain.Folio
	var notes *string
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, reservation_id, guest_id, property_id, folio_number,
		        status, subtotal, tax_total, total_amount, paid_amount, balance,
		        notes, created_at, updated_at, closed_at
		 FROM folios WHERE id = $1 AND tenant_id = $2`,
		folioID, tenantID,
	).Scan(&f.ID, &f.TenantID, &f.ReservationID, &f.GuestID, &f.PropertyID, &f.FolioNumber,
		&f.Status, &f.Subtotal, &f.TaxTotal, &f.TotalAmount, &f.PaidAmount, &f.Balance,
		&notes, &f.CreatedAt, &f.UpdatedAt, &f.ClosedAt)

	if err == pgx.ErrNoRows {
		return nil, apperrors.NotFound("folio", folioID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("get folio: %w", err)
	}
	if notes != nil {
		f.Notes = *notes
	}
	return &f, nil
}

// GetFolioByReservation returns the open folio for a reservation.
func (r *Repository) GetFolioByReservation(ctx context.Context, reservationID, tenantID uuid.UUID) (*domain.Folio, error) {
	var f domain.Folio
	var notes *string
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, reservation_id, guest_id, property_id, folio_number,
		        status, subtotal, tax_total, total_amount, paid_amount, balance,
		        notes, created_at, updated_at, closed_at
		 FROM folios WHERE reservation_id = $1 AND tenant_id = $2 AND status = 'open'`,
		reservationID, tenantID,
	).Scan(&f.ID, &f.TenantID, &f.ReservationID, &f.GuestID, &f.PropertyID, &f.FolioNumber,
		&f.Status, &f.Subtotal, &f.TaxTotal, &f.TotalAmount, &f.PaidAmount, &f.Balance,
		&notes, &f.CreatedAt, &f.UpdatedAt, &f.ClosedAt)

	if err == pgx.ErrNoRows {
		return nil, apperrors.NotFound("folio", "for reservation "+reservationID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("get folio by reservation: %w", err)
	}
	if notes != nil {
		f.Notes = *notes
	}
	return &f, nil
}

// AddLineItem adds a charge to a folio and recalculates totals.
func (r *Repository) AddLineItem(ctx context.Context, item *domain.LineItem) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Verify folio is open
	var folioStatus string
	err = tx.QueryRow(ctx,
		`SELECT status FROM folios WHERE id = $1 AND tenant_id = $2 FOR UPDATE`,
		item.FolioID, item.TenantID,
	).Scan(&folioStatus)
	if err == pgx.ErrNoRows {
		return apperrors.NotFound("folio", item.FolioID.String())
	}
	if err != nil {
		return fmt.Errorf("lock folio: %w", err)
	}
	if folioStatus != "open" {
		return apperrors.BadRequest("cannot add charges to a closed folio")
	}

	// Calculate amounts using decimal precision
	item.Amount = decimal.NewFromInt(int64(item.Quantity)).Mul(item.UnitPrice)
	item.TaxAmount = item.Amount.Mul(item.TaxRate).Div(decimal.NewFromInt(100)).Round(2)
	item.Total = item.Amount.Add(item.TaxAmount)

	// Insert line item
	err = tx.QueryRow(ctx,
		`INSERT INTO line_items (tenant_id, folio_id, category, description, quantity, unit_price, amount, tax_rate, tax_amount, total, date, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 RETURNING id, created_at`,
		item.TenantID, item.FolioID, item.Category, item.Description,
		item.Quantity, item.UnitPrice, item.Amount, item.TaxRate, item.TaxAmount, item.Total,
		item.Date, item.CreatedBy,
	).Scan(&item.ID, &item.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert line item: %w", err)
	}

	// Recalculate folio totals
	err = r.recalculateFolioTotals(ctx, tx, item.FolioID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// VoidLineItem voids a line item and recalculates folio.
func (r *Repository) VoidLineItem(ctx context.Context, itemID, tenantID uuid.UUID, reason string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var folioID uuid.UUID
	result, err := tx.Exec(ctx,
		`UPDATE line_items SET is_void = true, void_reason = $3
		 WHERE id = $1 AND tenant_id = $2 AND is_void = false`,
		itemID, tenantID, reason,
	)
	if err != nil {
		return fmt.Errorf("void line item: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.NotFound("line_item", itemID.String())
	}

	err = tx.QueryRow(ctx, `SELECT folio_id FROM line_items WHERE id = $1`, itemID).Scan(&folioID)
	if err != nil {
		return fmt.Errorf("get folio_id: %w", err)
	}

	if err := r.recalculateFolioTotals(ctx, tx, folioID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// ListLineItems returns all items for a folio.
func (r *Repository) ListLineItems(ctx context.Context, folioID, tenantID uuid.UUID) ([]domain.LineItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, folio_id, category, description, quantity, unit_price, amount,
		        tax_rate, tax_amount, total, date, is_void, COALESCE(void_reason, ''), created_at, COALESCE(created_by, '00000000-0000-0000-0000-000000000000')
		 FROM line_items WHERE folio_id = $1 AND tenant_id = $2 ORDER BY date, created_at`,
		folioID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("list line items: %w", err)
	}
	defer rows.Close()

	var items []domain.LineItem
	for rows.Next() {
		var item domain.LineItem
		if err := rows.Scan(&item.ID, &item.TenantID, &item.FolioID, &item.Category, &item.Description,
			&item.Quantity, &item.UnitPrice, &item.Amount, &item.TaxRate, &item.TaxAmount,
			&item.Total, &item.Date, &item.IsVoid, &item.VoidReason, &item.CreatedAt, &item.CreatedBy); err != nil {
			return nil, fmt.Errorf("scan line item: %w", err)
		}
		items = append(items, item)
	}
	if items == nil {
		items = []domain.LineItem{}
	}
	return items, nil
}

// RecordPayment records a payment and updates folio.
func (r *Repository) RecordPayment(ctx context.Context, payment *domain.Payment) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Insert payment
	err = tx.QueryRow(ctx,
		`INSERT INTO payments (tenant_id, folio_id, invoice_id, payment_type, payment_method, amount, reference_number, notes, received_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, created_at`,
		payment.TenantID, payment.FolioID, payment.InvoiceID,
		payment.PaymentType, payment.PaymentMethod, payment.Amount,
		payment.ReferenceNumber, payment.Notes, payment.ReceivedBy,
	).Scan(&payment.ID, &payment.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert payment: %w", err)
	}

	// Update folio paid amount and balance
	switch payment.PaymentType {
	case domain.PaymentTypePayment, domain.PaymentTypeDeposit:
		_, err = tx.Exec(ctx,
			`UPDATE folios SET paid_amount = paid_amount + $2, balance = balance - $2 WHERE id = $1 AND tenant_id = $3`,
			payment.FolioID, payment.Amount, payment.TenantID,
		)
	case domain.PaymentTypeRefund, domain.PaymentTypeDepositRelease:
		_, err = tx.Exec(ctx,
			`UPDATE folios SET paid_amount = paid_amount - $2, balance = balance + $2 WHERE id = $1 AND tenant_id = $3`,
			payment.FolioID, payment.Amount, payment.TenantID,
		)
	}
	if err != nil {
		return fmt.Errorf("update folio balance: %w", err)
	}

	// Update invoice status if linked
	if payment.InvoiceID != nil {
		var totalAmount, paidAmount decimal.Decimal
		// For refunds/releases, subtract from paid_amount; for payments/deposits, add to paid_amount
		switch payment.PaymentType {
		case domain.PaymentTypeRefund, domain.PaymentTypeDepositRelease:
			err = tx.QueryRow(ctx,
				`SELECT total_amount, paid_amount - $2 FROM invoices WHERE id = $1 AND tenant_id = $3`,
				payment.InvoiceID, payment.Amount, payment.TenantID,
			).Scan(&totalAmount, &paidAmount)
			if err == nil {
				status := "partially_paid"
				if paidAmount.LessThanOrEqual(decimal.Zero) {
					status = "refunded"
				} else if paidAmount.GreaterThanOrEqual(totalAmount) {
					status = "paid"
				}
				_, _ = tx.Exec(ctx,
					`UPDATE invoices SET paid_amount = paid_amount - $2, balance_due = balance_due + $2, status = $3 WHERE id = $1 AND tenant_id = $4`,
					payment.InvoiceID, payment.Amount, status, payment.TenantID,
				)
			}
		default:
			err = tx.QueryRow(ctx,
				`SELECT total_amount, paid_amount + $2 FROM invoices WHERE id = $1 AND tenant_id = $3`,
				payment.InvoiceID, payment.Amount, payment.TenantID,
			).Scan(&totalAmount, &paidAmount)
			if err == nil {
				status := "partially_paid"
				if paidAmount.GreaterThanOrEqual(totalAmount) {
					status = "paid"
				}
				_, _ = tx.Exec(ctx,
					`UPDATE invoices SET paid_amount = paid_amount + $2, balance_due = balance_due - $2, status = $3 WHERE id = $1 AND tenant_id = $4`,
					payment.InvoiceID, payment.Amount, status, payment.TenantID,
				)
			}
		}
	}

	return tx.Commit(ctx)
}

// PaymentExistsByReference checks if a payment with the given reference number already exists for a folio.
func (r *Repository) PaymentExistsByReference(ctx context.Context, tenantID, folioID uuid.UUID, referenceNumber string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM payments WHERE tenant_id = $1 AND folio_id = $2 AND reference_number = $3)`,
		tenantID, folioID, referenceNumber,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check payment exists: %w", err)
	}
	return exists, nil
}

// ListPayments returns all payments for a folio.
func (r *Repository) ListPayments(ctx context.Context, folioID, tenantID uuid.UUID) ([]domain.Payment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, folio_id, invoice_id, payment_type, payment_method, amount,
		        COALESCE(reference_number, ''), COALESCE(notes, ''), COALESCE(received_by, '00000000-0000-0000-0000-000000000000'), created_at
		 FROM payments WHERE folio_id = $1 AND tenant_id = $2 ORDER BY created_at`,
		folioID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("list payments: %w", err)
	}
	defer rows.Close()

	var payments []domain.Payment
	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(&p.ID, &p.TenantID, &p.FolioID, &p.InvoiceID, &p.PaymentType,
			&p.PaymentMethod, &p.Amount, &p.ReferenceNumber, &p.Notes, &p.ReceivedBy, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan payment: %w", err)
		}
		payments = append(payments, p)
	}
	if payments == nil {
		payments = []domain.Payment{}
	}
	return payments, nil
}

// GetInvoiceByID returns a full invoice with line items.
func (r *Repository) GetInvoiceByID(ctx context.Context, invoiceID, tenantID uuid.UUID) (*domain.Invoice, error) {
	var inv domain.Invoice
	var guestEmail, guestPhone, guestAddress, propGST, notes, pdfKey *string
	var dueDate *time.Time
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, folio_id, reservation_id, guest_id, property_id,
		        invoice_number, status, guest_name, guest_email, guest_phone, guest_address,
		        property_name, property_address, property_gst_number,
		        subtotal, cgst_amount, sgst_amount, igst_amount, total_tax, total_amount,
		        paid_amount, balance_due, check_in_date, check_out_date, num_nights,
		        issued_at, due_date, notes, pdf_key, created_at
		 FROM invoices WHERE id = $1 AND tenant_id = $2`,
		invoiceID, tenantID,
	).Scan(&inv.ID, &inv.TenantID, &inv.FolioID, &inv.ReservationID, &inv.GuestID, &inv.PropertyID,
		&inv.InvoiceNumber, &inv.Status, &inv.GuestName, &guestEmail, &guestPhone, &guestAddress,
		&inv.PropertyName, &inv.PropertyAddress, &propGST,
		&inv.Subtotal, &inv.CGSTAmount, &inv.SGSTAmount, &inv.IGSTAmount, &inv.TotalTax, &inv.TotalAmount,
		&inv.PaidAmount, &inv.BalanceDue, &inv.CheckInDate, &inv.CheckOutDate, &inv.NumNights,
		&inv.IssuedAt, &dueDate, &notes, &pdfKey, &inv.CreatedAt)

	if err == pgx.ErrNoRows {
		return nil, apperrors.NotFound("invoice", invoiceID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("get invoice: %w", err)
	}

	if guestEmail != nil {
		inv.GuestEmail = *guestEmail
	}
	if guestPhone != nil {
		inv.GuestPhone = *guestPhone
	}
	if guestAddress != nil {
		inv.GuestAddress = *guestAddress
	}
	if propGST != nil {
		inv.PropertyGST = *propGST
	}
	if notes != nil {
		inv.Notes = *notes
	}
	if pdfKey != nil {
		inv.PDFKey = *pdfKey
	}
	inv.DueDate = dueDate

	// Fetch line items
	rows, err := r.pool.Query(ctx,
		`SELECT id, invoice_id, category, description, quantity, unit_price, amount, tax_rate, tax_amount, total, date
		 FROM invoice_line_items WHERE invoice_id = $1 ORDER BY date`,
		invoiceID,
	)
	if err != nil {
		return nil, fmt.Errorf("get invoice line items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var li domain.InvoiceLineItem
		if err := rows.Scan(&li.ID, &li.InvoiceID, &li.Category, &li.Description,
			&li.Quantity, &li.UnitPrice, &li.Amount, &li.TaxRate, &li.TaxAmount, &li.Total, &li.Date); err != nil {
			return nil, fmt.Errorf("scan invoice line item: %w", err)
		}
		inv.LineItems = append(inv.LineItems, li)
	}
	if inv.LineItems == nil {
		inv.LineItems = []domain.InvoiceLineItem{}
	}

	return &inv, nil
}

// ListInvoices returns invoices for a tenant with optional filters.
func (r *Repository) ListInvoices(ctx context.Context, tenantID uuid.UUID, propertyID *uuid.UUID, status string, limit, offset int) ([]domain.Invoice, int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM invoices
		 WHERE tenant_id = $1
		   AND ($2::UUID IS NULL OR property_id = $2::UUID)
		   AND ($3::VARCHAR = '' OR status = $3::VARCHAR)`,
		tenantID, propertyID, status,
	).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("count invoices: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, folio_id, reservation_id, guest_id, property_id,
		        invoice_number, status, guest_name, guest_email, guest_phone, guest_address,
		        property_name, property_address, property_gst_number,
		        subtotal, cgst_amount, sgst_amount, igst_amount, total_tax, total_amount,
		        paid_amount, balance_due, check_in_date, check_out_date, num_nights,
		        issued_at, due_date, notes, pdf_key, created_at
		 FROM invoices
		 WHERE tenant_id = $1
		   AND ($2::UUID IS NULL OR property_id = $2::UUID)
		   AND ($3::VARCHAR = '' OR status = $3::VARCHAR)
		 ORDER BY issued_at DESC
		 LIMIT $4 OFFSET $5`,
		tenantID, propertyID, status, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list invoices: %w", err)
	}
	defer rows.Close()

	var invoices []domain.Invoice
	for rows.Next() {
		var inv domain.Invoice
		var guestEmail, guestPhone, guestAddress, propGST, notes, pdfKey *string
		var dueDate *time.Time
		if err := rows.Scan(&inv.ID, &inv.TenantID, &inv.FolioID, &inv.ReservationID, &inv.GuestID, &inv.PropertyID,
			&inv.InvoiceNumber, &inv.Status, &inv.GuestName, &guestEmail, &guestPhone, &guestAddress,
			&inv.PropertyName, &inv.PropertyAddress, &propGST,
			&inv.Subtotal, &inv.CGSTAmount, &inv.SGSTAmount, &inv.IGSTAmount, &inv.TotalTax, &inv.TotalAmount,
			&inv.PaidAmount, &inv.BalanceDue, &inv.CheckInDate, &inv.CheckOutDate, &inv.NumNights,
			&inv.IssuedAt, &dueDate, &notes, &pdfKey, &inv.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan invoice: %w", err)
		}
		if guestEmail != nil {
			inv.GuestEmail = *guestEmail
		}
		if guestPhone != nil {
			inv.GuestPhone = *guestPhone
		}
		if guestAddress != nil {
			inv.GuestAddress = *guestAddress
		}
		if propGST != nil {
			inv.PropertyGST = *propGST
		}
		if notes != nil {
			inv.Notes = *notes
		}
		if pdfKey != nil {
			inv.PDFKey = *pdfKey
		}
		inv.DueDate = dueDate
		invoices = append(invoices, inv)
	}
	if invoices == nil {
		invoices = []domain.Invoice{}
	}
	return invoices, count, nil
}

func (r *Repository) UpdateInvoicePDF(ctx context.Context, invoiceID, tenantID uuid.UUID, pdfKey string) error {
	_, err := r.pool.Exec(ctx, `UPDATE invoices SET pdf_key = $2 WHERE id = $1 AND tenant_id = $3`, invoiceID, pdfKey, tenantID)
	return err
}

// recalculateFolioTotals recalculates folio totals from non-voided line items.
func (r *Repository) recalculateFolioTotals(ctx context.Context, tx pgx.Tx, folioID uuid.UUID) error {
	var subtotal, taxTotal decimal.Decimal
	err := tx.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount), 0), COALESCE(SUM(tax_amount), 0)
		 FROM line_items WHERE folio_id = $1 AND is_void = false`,
		folioID,
	).Scan(&subtotal, &taxTotal)
	if err != nil {
		return fmt.Errorf("calc totals: %w", err)
	}

	totalAmount := subtotal.Add(taxTotal)

	var paidAmount decimal.Decimal
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(SUM(CASE WHEN payment_type IN ('payment', 'deposit') THEN amount ELSE -amount END), 0)
		 FROM payments WHERE folio_id = $1`,
		folioID,
	).Scan(&paidAmount)
	if err != nil {
		return fmt.Errorf("calc paid: %w", err)
	}

	balance := totalAmount.Sub(paidAmount)

	_, err = tx.Exec(ctx,
		`UPDATE folios SET subtotal = $2, tax_total = $3, total_amount = $4, paid_amount = $5, balance = $6 WHERE id = $1`,
		folioID, subtotal, taxTotal, totalAmount, paidAmount, balance,
	)
	if err != nil {
		return fmt.Errorf("update folio totals: %w", err)
	}

	return nil
}

// Refund records a refund payment.
func (r *Repository) Refund(ctx context.Context, payment *domain.Payment) error {
	return r.RecordPayment(ctx, payment)
}

// GetFolioSummary for generating invoices.
type FolioSummary struct {
	Folio    domain.Folio
	Items    []domain.LineItem
	Payments []domain.Payment
}

func (r *Repository) GetFolioSummary(ctx context.Context, folioID, tenantID uuid.UUID) (*FolioSummary, error) {
	folio, err := r.GetFolioByID(ctx, folioID, tenantID)
	if err != nil {
		return nil, err
	}

	items, err := r.ListLineItems(ctx, folioID, tenantID)
	if err != nil {
		return nil, err
	}

	payments, err := r.ListPayments(ctx, folioID, tenantID)
	if err != nil {
		return nil, err
	}

	return &FolioSummary{
		Folio:    *folio,
		Items:    items,
		Payments: payments,
	}, nil
}

// GetPaymentsByDateRange for reporting.
func (r *Repository) GetPaymentsByDateRange(ctx context.Context, tenantID uuid.UUID, start, end time.Time) ([]domain.Payment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, folio_id, invoice_id, payment_type, payment_method, amount,
		        COALESCE(reference_number, ''), COALESCE(notes, ''), COALESCE(received_by, '00000000-0000-0000-0000-000000000000'), created_at
		 FROM payments WHERE tenant_id = $1 AND created_at >= $2 AND created_at < $3
		 ORDER BY created_at`,
		tenantID, start, end,
	)
	if err != nil {
		return nil, fmt.Errorf("get payments by range: %w", err)
	}
	defer rows.Close()

	var payments []domain.Payment
	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(&p.ID, &p.TenantID, &p.FolioID, &p.InvoiceID, &p.PaymentType,
			&p.PaymentMethod, &p.Amount, &p.ReferenceNumber, &p.Notes, &p.ReceivedBy, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan payment: %w", err)
		}
		payments = append(payments, p)
	}
	if payments == nil {
		payments = []domain.Payment{}
	}
	return payments, nil
}
