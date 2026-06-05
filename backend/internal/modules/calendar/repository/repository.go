package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stayflow/stayflow-track/internal/modules/calendar/domain"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetCalendarView fetches all units and their reservations for a date range.
func (r *Repository) GetCalendarView(ctx context.Context, tenantID, propertyID uuid.UUID, start, end time.Time) (*domain.CalendarView, error) {
	var propertyName string
	err := r.pool.QueryRow(ctx,
		`SELECT name FROM properties WHERE id = $1 AND tenant_id = $2`,
		propertyID, tenantID,
	).Scan(&propertyName)
	if err == pgx.ErrNoRows {
		return nil, apperrors.NotFound("property", propertyID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("get property name: %w", err)
	}

	// Fetch all units for the property
	unitRows, err := r.pool.Query(ctx,
		`SELECT u.id, u.unit_number, u.floor, u.status, ut.name, ut.base_rate
		 FROM units u
		 JOIN unit_types ut ON u.unit_type_id = ut.id
		 WHERE u.property_id = $1 AND u.tenant_id = $2
		 ORDER BY u.floor, u.unit_number`,
		propertyID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("fetch units: %w", err)
	}
	defer unitRows.Close()

	var units []domain.CalendarUnit
	unitMap := make(map[uuid.UUID]int) // unit_id -> index

	for unitRows.Next() {
		var u domain.CalendarUnit
		if err := unitRows.Scan(&u.UnitID, &u.UnitNumber, &u.Floor, &u.Status, &u.UnitTypeName, &u.BaseRate); err != nil {
			return nil, fmt.Errorf("scan unit: %w", err)
		}
		u.Entries = []domain.CalendarEntry{}
		unitMap[u.UnitID] = len(units)
		units = append(units, u)
	}

	// Fetch reservations overlapping with the date range
	resRows, err := r.pool.Query(ctx,
		`SELECT r.id, r.unit_id, r.guest_id,
		        g.first_name || ' ' || g.last_name AS guest_name,
		        r.check_in_date, r.check_out_date, r.status, r.booking_source
		 FROM reservations r
		 JOIN guests g ON r.guest_id = g.id
		 WHERE r.property_id = $1
		   AND r.tenant_id = $2
		   AND r.status NOT IN ('cancelled')
		   AND r.check_in_date < $4
		   AND r.check_out_date > $3
		 ORDER BY r.check_in_date`,
		propertyID, tenantID, start, end,
	)
	if err != nil {
		return nil, fmt.Errorf("fetch reservations: %w", err)
	}
	defer resRows.Close()

	for resRows.Next() {
		var entry domain.CalendarEntry
		if err := resRows.Scan(&entry.ReservationID, &entry.UnitID, &entry.GuestID,
			&entry.GuestName, &entry.CheckInDate, &entry.CheckOutDate,
			&entry.Status, &entry.BookingSource); err != nil {
			return nil, fmt.Errorf("scan reservation: %w", err)
		}
		entry.NumNights = int(entry.CheckOutDate.Sub(entry.CheckInDate).Hours() / 24)

		if idx, ok := unitMap[entry.UnitID]; ok {
			units[idx].Entries = append(units[idx].Entries, entry)
		}
	}

	return &domain.CalendarView{
		PropertyID:   propertyID,
		PropertyName: propertyName,
		StartDate:    start,
		EndDate:      end,
		Units:        units,
	}, nil
}

// MoveReservation updates unit and dates within a transaction.
func (r *Repository) MoveReservation(ctx context.Context, tenantID, reservationID, newUnitID uuid.UUID, newCheckIn, newCheckOut time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check conflict for new position
	var hasConflict bool
	err = tx.QueryRow(ctx,
		`SELECT check_reservation_conflict($1, $2, $3, $4)`,
		newUnitID, newCheckIn, newCheckOut, reservationID,
	).Scan(&hasConflict)
	if err != nil {
		return fmt.Errorf("check conflict: %w", err)
	}
	if hasConflict {
		return apperrors.Conflict("unit is not available for the selected dates")
	}

	// Get current reservation to release old unit
	var oldUnitID uuid.UUID
	var currentStatus string
	err = tx.QueryRow(ctx,
		`SELECT unit_id, status FROM reservations WHERE id = $1 AND tenant_id = $2`,
		reservationID, tenantID,
	).Scan(&oldUnitID, &currentStatus)
	if err == pgx.ErrNoRows {
		return apperrors.NotFound("reservation", reservationID.String())
	}
	if err != nil {
		return fmt.Errorf("get reservation: %w", err)
	}

	if currentStatus == "cancelled" || currentStatus == "checked_out" {
		return apperrors.BadRequest("cannot move a cancelled or checked-out reservation")
	}

	// Calculate new total
	nights := int(newCheckOut.Sub(newCheckIn).Hours() / 24)
	var ratePerNight float64
	err = tx.QueryRow(ctx,
		`SELECT rate_per_night FROM reservations WHERE id = $1`, reservationID,
	).Scan(&ratePerNight)
	if err != nil {
		return fmt.Errorf("get rate: %w", err)
	}
	newTotal := float64(nights) * ratePerNight

	// Update reservation
	_, err = tx.Exec(ctx,
		`UPDATE reservations SET
			unit_id = $3, check_in_date = $4, check_out_date = $5, total_amount = $6
		 WHERE id = $1 AND tenant_id = $2`,
		reservationID, tenantID, newUnitID, newCheckIn, newCheckOut, newTotal,
	)
	if err != nil {
		return fmt.Errorf("update reservation: %w", err)
	}

	// Update unit statuses if checked_in
	if currentStatus == "checked_in" {
		if oldUnitID != newUnitID {
			_, err = tx.Exec(ctx, `UPDATE units SET status = 'available' WHERE id = $1`, oldUnitID)
			if err != nil {
				return fmt.Errorf("release old unit: %w", err)
			}
			_, err = tx.Exec(ctx, `UPDATE units SET status = 'occupied' WHERE id = $1`, newUnitID)
			if err != nil {
				return fmt.Errorf("occupy new unit: %w", err)
			}
		}
	}

	return tx.Commit(ctx)
}

// GetOccupancyStats returns daily occupancy for a date range.
func (r *Repository) GetOccupancyStats(ctx context.Context, tenantID, propertyID uuid.UUID, start, end time.Time) ([]domain.OccupancyStats, error) {
	var totalUnits int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM units WHERE property_id = $1 AND tenant_id = $2`,
		propertyID, tenantID,
	).Scan(&totalUnits)
	if err != nil {
		return nil, fmt.Errorf("count units: %w", err)
	}

	var stats []domain.OccupancyStats
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		nextDay := d.AddDate(0, 0, 1)
		var occupied int
		err := r.pool.QueryRow(ctx,
			`SELECT COUNT(DISTINCT unit_id) FROM reservations
			 WHERE property_id = $1 AND tenant_id = $2
			   AND status NOT IN ('cancelled', 'checked_out')
			   AND check_in_date < $4 AND check_out_date > $3`,
			propertyID, tenantID, d, nextDay,
		).Scan(&occupied)
		if err != nil {
			return nil, fmt.Errorf("count occupied: %w", err)
		}

		rate := 0.0
		if totalUnits > 0 {
			rate = float64(occupied) / float64(totalUnits) * 100
		}

		stats = append(stats, domain.OccupancyStats{
			Date:           d,
			TotalUnits:     totalUnits,
			OccupiedUnits:  occupied,
			AvailableUnits: totalUnits - occupied,
			OccupancyRate:  rate,
		})
	}

	return stats, nil
}
