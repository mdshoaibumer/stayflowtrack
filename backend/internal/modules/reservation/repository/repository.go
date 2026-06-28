package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stayflow/stayflow-track/internal/modules/reservation/domain"
	"github.com/stayflow/stayflow-track/internal/platform/database"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

type Repository struct {
	pool *database.TenantPool
}

func New(pool *database.TenantPool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateReservation(ctx context.Context, res *domain.Reservation) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO reservations (
			tenant_id, property_id, unit_id, guest_id,
			booking_source, status, check_in_date, check_out_date,
			num_guests, rate_per_night, total_amount, notes, external_booking_id,
			advance_amount, advance_method, advance_reference
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, created_at, updated_at`,
		res.TenantID, res.PropertyID, res.UnitID, res.GuestID,
		res.BookingSource, res.Status, res.CheckInDate, res.CheckOutDate,
		res.NumGuests, res.RatePerNight, res.TotalAmount, res.Notes, res.ExternalBookingID,
		res.AdvanceAmount, res.AdvanceMethod, res.AdvanceReference,
	).Scan(&res.ID, &res.CreatedAt, &res.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create reservation: %w", err)
	}
	return nil
}

func (r *Repository) GetReservationByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Reservation, error) {
	var res domain.Reservation
	err := r.pool.QueryRow(ctx,
		`SELECT r.id, r.tenant_id, r.property_id, r.unit_id, r.guest_id,
		        r.booking_source, r.status, r.check_in_date, r.check_out_date,
		        r.actual_check_in, r.actual_check_out, r.num_guests,
		        r.rate_per_night, r.total_amount, COALESCE(r.notes, ''),
		        COALESCE(r.cancellation_reason, ''), r.cancelled_at, COALESCE(r.external_booking_id, ''),
		        r.created_at, r.updated_at,
		        g.first_name, g.last_name, COALESCE(g.phone, ''),
		        u.unit_number, ut.name, p.name,
		        COALESCE(r.advance_amount, 0), COALESCE(r.advance_method, ''), COALESCE(r.advance_reference, '')
		 FROM reservations r
		 JOIN guests g ON r.guest_id = g.id
		 JOIN units u ON r.unit_id = u.id
		 JOIN unit_types ut ON u.unit_type_id = ut.id
		 JOIN properties p ON r.property_id = p.id
		 WHERE r.id = $1 AND r.tenant_id = $2`, id, tenantID,
	).Scan(&res.ID, &res.TenantID, &res.PropertyID, &res.UnitID, &res.GuestID,
		&res.BookingSource, &res.Status, &res.CheckInDate, &res.CheckOutDate,
		&res.ActualCheckIn, &res.ActualCheckOut, &res.NumGuests,
		&res.RatePerNight, &res.TotalAmount, &res.Notes,
		&res.CancellationReason, &res.CancelledAt, &res.ExternalBookingID,
		&res.CreatedAt, &res.UpdatedAt,
		&res.GuestFirstName, &res.GuestLastName, &res.GuestPhone,
		&res.UnitNumber, &res.UnitTypeName, &res.PropertyName,
		&res.AdvanceAmount, &res.AdvanceMethod, &res.AdvanceReference)

	if err == pgx.ErrNoRows {
		return nil, apperrors.NotFound("reservation", id.String())
	}
	if err != nil {
		return nil, fmt.Errorf("get reservation: %w", err)
	}
	return &res, nil
}

type ListParams struct {
	TenantID   uuid.UUID
	PropertyID *uuid.UUID
	Status     string
	Limit      int
	Offset     int
}

func (r *Repository) ListReservations(ctx context.Context, params ListParams) ([]domain.Reservation, int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM reservations
		 WHERE tenant_id = $1
		   AND ($2::UUID IS NULL OR property_id = $2::UUID)
		   AND ($3::VARCHAR = '' OR status = $3::VARCHAR)`,
		params.TenantID, params.PropertyID, params.Status,
	).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("count reservations: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT r.id, r.tenant_id, r.property_id, r.unit_id, r.guest_id,
		        r.booking_source, r.status, r.check_in_date, r.check_out_date,
		        r.actual_check_in, r.actual_check_out, r.num_guests,
		        r.rate_per_night, r.total_amount, COALESCE(r.notes, ''),
		        COALESCE(r.cancellation_reason, ''), r.cancelled_at, COALESCE(r.external_booking_id, ''),
		        r.created_at, r.updated_at,
		        g.first_name, g.last_name, COALESCE(g.phone, ''),
		        u.unit_number, ut.name,
		        COALESCE(r.advance_amount, 0), COALESCE(r.advance_method, ''), COALESCE(r.advance_reference, '')
		 FROM reservations r
		 JOIN guests g ON r.guest_id = g.id
		 JOIN units u ON r.unit_id = u.id
		 JOIN unit_types ut ON u.unit_type_id = ut.id
		 WHERE r.tenant_id = $1
		   AND ($2::UUID IS NULL OR r.property_id = $2::UUID)
		   AND ($3::VARCHAR = '' OR r.status = $3::VARCHAR)
		 ORDER BY r.check_in_date DESC
		 LIMIT $4 OFFSET $5`,
		params.TenantID, params.PropertyID, params.Status, params.Limit, params.Offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list reservations: %w", err)
	}
	defer rows.Close()

	var reservations []domain.Reservation
	for rows.Next() {
		var res domain.Reservation
		if err := rows.Scan(&res.ID, &res.TenantID, &res.PropertyID, &res.UnitID, &res.GuestID,
			&res.BookingSource, &res.Status, &res.CheckInDate, &res.CheckOutDate,
			&res.ActualCheckIn, &res.ActualCheckOut, &res.NumGuests,
			&res.RatePerNight, &res.TotalAmount, &res.Notes,
			&res.CancellationReason, &res.CancelledAt, &res.ExternalBookingID,
			&res.CreatedAt, &res.UpdatedAt,
			&res.GuestFirstName, &res.GuestLastName, &res.GuestPhone,
			&res.UnitNumber, &res.UnitTypeName,
			&res.AdvanceAmount, &res.AdvanceMethod, &res.AdvanceReference); err != nil {
			return nil, 0, fmt.Errorf("scan reservation: %w", err)
		}
		reservations = append(reservations, res)
	}

	if reservations == nil {
		reservations = []domain.Reservation{}
	}

	return reservations, count, nil
}

func (r *Repository) UpdateReservation(ctx context.Context, res *domain.Reservation) error {
	err := r.pool.QueryRow(ctx,
		`UPDATE reservations SET
			check_in_date = $3,
			check_out_date = $4,
			num_guests = $5,
			rate_per_night = $6,
			total_amount = $7,
			notes = $8
		 WHERE id = $1 AND tenant_id = $2 AND status IN ('pending', 'confirmed')
		 RETURNING id, tenant_id, property_id, unit_id, guest_id, booking_source, status,
		           check_in_date, check_out_date, num_guests, rate_per_night, total_amount,
		           notes, created_at, updated_at`,
		res.ID, res.TenantID, res.CheckInDate, res.CheckOutDate,
		res.NumGuests, res.RatePerNight, res.TotalAmount, res.Notes,
	).Scan(&res.ID, &res.TenantID, &res.PropertyID, &res.UnitID, &res.GuestID,
		&res.BookingSource, &res.Status, &res.CheckInDate, &res.CheckOutDate,
		&res.NumGuests, &res.RatePerNight, &res.TotalAmount, &res.Notes,
		&res.CreatedAt, &res.UpdatedAt)

	if err == pgx.ErrNoRows {
		return apperrors.NotFound("reservation", res.ID.String())
	}
	if err != nil {
		return fmt.Errorf("update reservation: %w", err)
	}
	return nil
}

func (r *Repository) CancelReservation(ctx context.Context, id, tenantID uuid.UUID, reason string) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE reservations SET
			status = 'cancelled',
			cancellation_reason = $3,
			cancelled_at = NOW()
		 WHERE id = $1 AND tenant_id = $2 AND status IN ('pending', 'confirmed')`,
		id, tenantID, reason,
	)
	if err != nil {
		return fmt.Errorf("cancel reservation: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.BadRequest("reservation cannot be cancelled in its current state")
	}
	return nil
}

func (r *Repository) ConfirmReservation(ctx context.Context, id, tenantID uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE reservations SET status = 'confirmed', updated_at = NOW()
		 WHERE id = $1 AND tenant_id = $2 AND status = 'pending'`,
		id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("confirm reservation: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.BadRequest("reservation must be in 'pending' status to confirm")
	}
	return nil
}

func (r *Repository) CheckIn(ctx context.Context, id, tenantID uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE reservations SET status = 'checked_in', actual_check_in = NOW()
		 WHERE id = $1 AND tenant_id = $2 AND status = 'confirmed'`,
		id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("check in: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.BadRequest("reservation must be in 'confirmed' status to check in")
	}
	return nil
}

func (r *Repository) CheckOut(ctx context.Context, id, tenantID uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE reservations SET status = 'checked_out', actual_check_out = NOW()
		 WHERE id = $1 AND tenant_id = $2 AND status = 'checked_in'`,
		id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("check out: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.BadRequest("reservation must be in 'checked_in' status to check out")
	}
	return nil
}

func (r *Repository) CheckConflict(ctx context.Context, unitID uuid.UUID, checkIn, checkOut time.Time, excludeID *uuid.UUID) (bool, error) {
	var hasConflict bool
	err := r.pool.QueryRow(ctx,
		`SELECT check_reservation_conflict($1, $2, $3, $4)`,
		unitID, checkIn, checkOut, excludeID,
	).Scan(&hasConflict)
	if err != nil {
		return false, fmt.Errorf("check conflict: %w", err)
	}
	return hasConflict, nil
}

// CheckConflictForUpdate performs a conflict check within a serializable-like advisory lock scope.
func (r *Repository) CheckConflictForUpdate(ctx context.Context, unitID uuid.UUID, checkIn, checkOut time.Time, excludeID *uuid.UUID, tenantID uuid.UUID) (bool, error) {
	var hasConflict bool
	err := r.pool.QueryRow(ctx,
		`SELECT check_reservation_conflict($1, $2, $3, $4)`,
		unitID, checkIn, checkOut, excludeID,
	).Scan(&hasConflict)
	if err != nil {
		return false, fmt.Errorf("check conflict for update: %w", err)
	}
	return hasConflict, nil
}

// CreateReservationAtomic performs conflict check + insert in a single transaction to prevent double-booking.
func (r *Repository) CreateReservationAtomic(ctx context.Context, res *domain.Reservation) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Acquire advisory lock on the unit to serialize concurrent booking attempts
	_, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1::text))`, res.UnitID)
	if err != nil {
		return fmt.Errorf("acquire advisory lock: %w", err)
	}

	// Check for conflicts within the transaction
	var hasConflict bool
	err = tx.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM reservations
			WHERE unit_id = $1
			  AND status NOT IN ('cancelled', 'checked_out')
			  AND check_in_date < $3
			  AND check_out_date > $2
		)`,
		res.UnitID, res.CheckInDate, res.CheckOutDate,
	).Scan(&hasConflict)
	if err != nil {
		return fmt.Errorf("check conflict in tx: %w", err)
	}
	if hasConflict {
		return apperrors.Conflict("unit is not available for the selected dates")
	}

	// Insert reservation
	err = tx.QueryRow(ctx,
		`INSERT INTO reservations (
			tenant_id, property_id, unit_id, guest_id,
			booking_source, status, check_in_date, check_out_date,
			num_guests, rate_per_night, total_amount, notes, external_booking_id,
			advance_amount, advance_method, advance_reference
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, created_at, updated_at`,
		res.TenantID, res.PropertyID, res.UnitID, res.GuestID,
		res.BookingSource, res.Status, res.CheckInDate, res.CheckOutDate,
		res.NumGuests, res.RatePerNight, res.TotalAmount, res.Notes, res.ExternalBookingID,
		res.AdvanceAmount, res.AdvanceMethod, res.AdvanceReference,
	).Scan(&res.ID, &res.CreatedAt, &res.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create reservation in tx: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *Repository) GetAvailableUnits(ctx context.Context, propertyID, tenantID uuid.UUID, checkIn, checkOut time.Time) ([]domain.AvailableUnit, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT u.id, u.unit_number, COALESCE(u.floor, ''), ut.name, ut.base_rate
		 FROM units u
		 JOIN unit_types ut ON u.unit_type_id = ut.id
		 WHERE u.property_id = $1
		   AND u.tenant_id = $2
		   AND u.status = 'available'
		   AND u.id NOT IN (
		       SELECT unit_id FROM reservations
		       WHERE property_id = $1
		         AND tenant_id = $2
		         AND status NOT IN ('cancelled', 'checked_out')
		         AND check_in_date < $4
		         AND check_out_date > $3
		   )
		 ORDER BY u.unit_number`,
		propertyID, tenantID, checkIn, checkOut,
	)
	if err != nil {
		return nil, fmt.Errorf("get available units: %w", err)
	}
	defer rows.Close()

	var units []domain.AvailableUnit
	for rows.Next() {
		var u domain.AvailableUnit
		if err := rows.Scan(&u.ID, &u.UnitNumber, &u.Floor, &u.UnitTypeName, &u.BaseRate); err != nil {
			return nil, fmt.Errorf("scan available unit: %w", err)
		}
		units = append(units, u)
	}

	if units == nil {
		units = []domain.AvailableUnit{}
	}

	return units, nil
}

func (r *Repository) UpdateUnitStatus(ctx context.Context, unitID uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE units SET status = $2 WHERE id = $1`, unitID, status)
	return err
}

// UpdateUnitStatusWithTenant updates unit status with tenant isolation.
func (r *Repository) UpdateUnitStatusWithTenant(ctx context.Context, unitID, tenantID uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE units SET status = $2 WHERE id = $1 AND tenant_id = $3`, unitID, status, tenantID)
	return err
}

func (r *Repository) GetReservationGuestID(ctx context.Context, id, tenantID uuid.UUID) (uuid.UUID, error) {
	var guestID uuid.UUID
	err := r.pool.QueryRow(ctx,
		`SELECT guest_id FROM reservations WHERE id = $1 AND tenant_id = $2`, id, tenantID,
	).Scan(&guestID)
	if err == pgx.ErrNoRows {
		return uuid.Nil, apperrors.NotFound("reservation", id.String())
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("get reservation guest: %w", err)
	}
	return guestID, nil
}

func (r *Repository) IncrementGuestStays(ctx context.Context, guestID, tenantID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE guests SET total_stays = total_stays + 1 WHERE id = $1 AND tenant_id = $2`, guestID, tenantID)
	return err
}
