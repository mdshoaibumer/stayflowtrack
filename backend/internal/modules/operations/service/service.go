package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/stayflow/stayflow-track/internal/modules/operations/domain"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

// Service handles extended hospitality operations.
type Service struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// MarkNoShow marks a confirmed reservation as no-show and optionally charges a fee.
func (s *Service) MarkNoShow(ctx context.Context, tenantID uuid.UUID, input domain.NoShowInput) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	result, err := tx.Exec(ctx,
		`UPDATE reservations SET
			is_no_show = true,
			no_show_at = NOW(),
			no_show_charge = $3,
			status = 'cancelled',
			cancellation_reason = 'no_show',
			cancelled_at = NOW()
		 WHERE id = $1 AND tenant_id = $2 AND status = 'confirmed'`,
		input.ReservationID, tenantID, input.ChargeAmount,
	)
	if err != nil {
		return fmt.Errorf("mark no-show: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.BadRequest("reservation must be in 'confirmed' status to mark as no-show")
	}

	// Release the unit back to available
	_, err = tx.Exec(ctx,
		`UPDATE units SET status = 'available'
		 WHERE id = (SELECT unit_id FROM reservations WHERE id = $1) AND tenant_id = $2`,
		input.ReservationID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("release unit: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// ExtendStay extends a checked-in guest's stay, checking availability and adding charges.
func (s *Service) ExtendStay(ctx context.Context, tenantID, userID uuid.UUID, input domain.ExtendStayInput) (*domain.StayExtension, error) {
	newCheckOut, err := time.Parse("2006-01-02", input.NewCheckOut)
	if err != nil {
		return nil, apperrors.BadRequest("invalid new_check_out format, use YYYY-MM-DD")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Get current reservation
	var unitID uuid.UUID
	var currentCheckOut time.Time
	var ratePerNight decimal.Decimal
	err = tx.QueryRow(ctx,
		`SELECT unit_id, check_out_date, rate_per_night
		 FROM reservations WHERE id = $1 AND tenant_id = $2 AND status = 'checked_in'`,
		input.ReservationID, tenantID,
	).Scan(&unitID, &currentCheckOut, &ratePerNight)
	if err != nil {
		return nil, apperrors.BadRequest("reservation must be in 'checked_in' status to extend")
	}

	if !newCheckOut.After(currentCheckOut) {
		return nil, apperrors.BadRequest("new check-out date must be after current check-out date")
	}

	// Check for conflicts with other reservations
	var hasConflict bool
	err = tx.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM reservations
			WHERE unit_id = $1 AND id != $2
			  AND status NOT IN ('cancelled', 'checked_out')
			  AND check_in_date < $3 AND check_out_date > $4
		)`,
		unitID, input.ReservationID, newCheckOut, currentCheckOut,
	).Scan(&hasConflict)
	if err != nil {
		return nil, fmt.Errorf("check conflict: %w", err)
	}
	if hasConflict {
		return nil, apperrors.Conflict("unit is not available for the extended dates")
	}

	// Calculate additional charges
	rate := input.RatePerNight
	if rate.IsZero() {
		rate = ratePerNight
	}
	additionalNights := int(newCheckOut.Sub(currentCheckOut).Hours() / 24)
	additionalAmount := decimal.NewFromInt(int64(additionalNights)).Mul(rate)

	// Update reservation
	_, err = tx.Exec(ctx,
		`UPDATE reservations SET check_out_date = $3, total_amount = total_amount + $4
		 WHERE id = $1 AND tenant_id = $2`,
		input.ReservationID, tenantID, newCheckOut, additionalAmount,
	)
	if err != nil {
		return nil, fmt.Errorf("update reservation: %w", err)
	}

	// Record extension
	ext := &domain.StayExtension{
		TenantID:         tenantID,
		ReservationID:    input.ReservationID,
		OriginalCheckOut: currentCheckOut,
		NewCheckOut:      newCheckOut,
		AdditionalNights: additionalNights,
		RatePerNight:     rate,
		AdditionalAmount: additionalAmount,
		Reason:           input.Reason,
		ExtendedBy:       userID,
	}

	err = tx.QueryRow(ctx,
		`INSERT INTO stay_extensions (tenant_id, reservation_id, original_check_out, new_check_out,
			additional_nights, rate_per_night, additional_amount, reason, extended_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, created_at`,
		ext.TenantID, ext.ReservationID, ext.OriginalCheckOut, ext.NewCheckOut,
		ext.AdditionalNights, ext.RatePerNight, ext.AdditionalAmount, ext.Reason, ext.ExtendedBy,
	).Scan(&ext.ID, &ext.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("record extension: %w", err)
	}

	// Auto-post extension charges to the guest's folio
	var folioID uuid.UUID
	err = tx.QueryRow(ctx,
		`SELECT id FROM folios WHERE reservation_id = $1 AND tenant_id = $2 AND status = 'open' LIMIT 1`,
		input.ReservationID, tenantID,
	).Scan(&folioID)
	if err == nil {
		// Calculate tax (18% GST on room charges)
		taxRate := decimal.NewFromInt(18)
		taxAmount := additionalAmount.Mul(taxRate).Div(decimal.NewFromInt(100)).Round(2)
		totalWithTax := additionalAmount.Add(taxAmount)

		description := fmt.Sprintf("Stay extension: %d additional night(s) @ ₹%s/night", additionalNights, rate.StringFixed(0))
		_, err = tx.Exec(ctx,
			`INSERT INTO folio_line_items (tenant_id, folio_id, category, description, quantity, unit_price, tax_rate, tax_amount, amount, date, created_by)
			 VALUES ($1, $2, 'room_charge', $3, $4, $5, $6, $7, $8, NOW(), $9)`,
			tenantID, folioID, description, additionalNights, rate, taxRate, taxAmount, totalWithTax, userID,
		)
		if err != nil {
			return nil, fmt.Errorf("post extension charge to folio: %w", err)
		}

		// Recalculate folio totals
		_, err = tx.Exec(ctx,
			`UPDATE folios SET
				total_charges = (SELECT COALESCE(SUM(amount), 0) FROM folio_line_items WHERE folio_id = $1 AND voided = false),
				balance = (SELECT COALESCE(SUM(amount), 0) FROM folio_line_items WHERE folio_id = $1 AND voided = false) -
						  (SELECT COALESCE(SUM(amount), 0) FROM payments WHERE folio_id = $1),
				updated_at = NOW()
			 WHERE id = $1`,
			folioID,
		)
		if err != nil {
			return nil, fmt.Errorf("recalculate folio: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return ext, nil
}

// MoveRoom moves a checked-in guest to a different unit.
func (s *Service) MoveRoom(ctx context.Context, tenantID, userID uuid.UUID, input domain.RoomMoveInput) (*domain.RoomMove, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Get current unit
	var fromUnitID uuid.UUID
	var checkOut time.Time
	err = tx.QueryRow(ctx,
		`SELECT unit_id, check_out_date FROM reservations
		 WHERE id = $1 AND tenant_id = $2 AND status = 'checked_in'`,
		input.ReservationID, tenantID,
	).Scan(&fromUnitID, &checkOut)
	if err != nil {
		return nil, apperrors.BadRequest("reservation must be in 'checked_in' status to move rooms")
	}

	// Check new unit is available
	var newUnitAvailable bool
	err = tx.QueryRow(ctx,
		`SELECT status = 'available' FROM units WHERE id = $1 AND tenant_id = $2`,
		input.ToUnitID, tenantID,
	).Scan(&newUnitAvailable)
	if err != nil || !newUnitAvailable {
		return nil, apperrors.BadRequest("target unit is not available")
	}

	// Update reservation to new unit
	_, err = tx.Exec(ctx,
		`UPDATE reservations SET unit_id = $3 WHERE id = $1 AND tenant_id = $2`,
		input.ReservationID, tenantID, input.ToUnitID,
	)
	if err != nil {
		return nil, fmt.Errorf("update reservation unit: %w", err)
	}

	// Release old unit, occupy new unit (include tenant_id for defense-in-depth)
	_, err = tx.Exec(ctx, `UPDATE units SET status = 'cleaning' WHERE id = $1 AND tenant_id = $2`, fromUnitID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("release old unit: %w", err)
	}
	_, err = tx.Exec(ctx, `UPDATE units SET status = 'occupied' WHERE id = $1 AND tenant_id = $2`, input.ToUnitID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("occupy new unit: %w", err)
	}

	// Record move
	move := &domain.RoomMove{
		TenantID:      tenantID,
		ReservationID: input.ReservationID,
		FromUnitID:    fromUnitID,
		ToUnitID:      input.ToUnitID,
		Reason:        input.Reason,
		RateChange:    input.RateChange,
		MovedBy:       userID,
	}

	err = tx.QueryRow(ctx,
		`INSERT INTO room_moves (tenant_id, reservation_id, from_unit_id, to_unit_id, reason, rate_change, moved_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, moved_at`,
		move.TenantID, move.ReservationID, move.FromUnitID, move.ToUnitID,
		move.Reason, move.RateChange, move.MovedBy,
	).Scan(&move.ID, &move.MovedAt)
	if err != nil {
		return nil, fmt.Errorf("record move: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return move, nil
}

// CreateMaintenanceBlock marks a unit as unavailable for a date range.
func (s *Service) CreateMaintenanceBlock(ctx context.Context, tenantID, userID uuid.UUID, input domain.MaintenanceBlockInput) (*domain.MaintenanceBlock, error) {
	startDate, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		return nil, apperrors.BadRequest("invalid start_date format")
	}
	endDate, err := time.Parse("2006-01-02", input.EndDate)
	if err != nil {
		return nil, apperrors.BadRequest("invalid end_date format")
	}
	if endDate.Before(startDate) {
		return nil, apperrors.BadRequest("end_date must be on or after start_date")
	}

	// Check for conflicts
	var hasConflict bool
	err = s.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM reservations
			WHERE unit_id = $1 AND tenant_id = $2
			  AND status NOT IN ('cancelled', 'checked_out')
			  AND check_in_date < $4 AND check_out_date > $3
		)`,
		input.UnitID, tenantID, startDate, endDate,
	).Scan(&hasConflict)
	if err != nil {
		return nil, fmt.Errorf("check conflicts: %w", err)
	}
	if hasConflict {
		return nil, apperrors.Conflict("unit has active reservations during the maintenance period")
	}

	block := &domain.MaintenanceBlock{
		TenantID:   tenantID,
		PropertyID: input.PropertyID,
		UnitID:     input.UnitID,
		Reason:     input.Reason,
		BlockType:  input.BlockType,
		StartDate:  startDate,
		EndDate:    endDate,
		CreatedBy:  userID,
		Notes:      input.Notes,
	}

	err = s.pool.QueryRow(ctx,
		`INSERT INTO maintenance_blocks (tenant_id, property_id, unit_id, reason, block_type, start_date, end_date, created_by, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, created_at, updated_at`,
		block.TenantID, block.PropertyID, block.UnitID, block.Reason, block.BlockType,
		block.StartDate, block.EndDate, block.CreatedBy, block.Notes,
	).Scan(&block.ID, &block.CreatedAt, &block.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create maintenance block: %w", err)
	}

	// If block starts today or earlier, mark unit as maintenance
	if !startDate.After(time.Now()) {
		_, _ = s.pool.Exec(ctx,
			`UPDATE units SET status = 'maintenance' WHERE id = $1 AND tenant_id = $2`,
			input.UnitID, tenantID,
		)
	}

	return block, nil
}

// RefundDeposit processes a deposit refund during check-out.
func (s *Service) RefundDeposit(ctx context.Context, tenantID, userID uuid.UUID, input domain.RefundDepositInput) error {
	result, err := s.pool.Exec(ctx,
		`UPDATE deposits SET
			status = 'refunded',
			released_at = NOW(),
			released_by = $3,
			refund_amount = $4,
			refund_method = $5,
			refund_reference = $6,
			notes = COALESCE($7, notes)
		 WHERE id = $1 AND tenant_id = $2 AND status = 'held'`,
		input.DepositID, tenantID, userID, input.RefundAmount,
		input.RefundMethod, input.RefundReference, input.Notes,
	)
	if err != nil {
		return fmt.Errorf("refund deposit: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.BadRequest("deposit not found or already processed")
	}
	return nil
}
