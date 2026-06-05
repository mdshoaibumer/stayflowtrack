package domain

import (
	"time"

	"github.com/google/uuid"
)

// CalendarEntry represents a single reservation block on the calendar grid.
type CalendarEntry struct {
	ReservationID uuid.UUID `json:"reservation_id"`
	UnitID        uuid.UUID `json:"unit_id"`
	GuestID       uuid.UUID `json:"guest_id"`
	GuestName     string    `json:"guest_name"`
	CheckInDate   time.Time `json:"check_in_date"`
	CheckOutDate  time.Time `json:"check_out_date"`
	Status        string    `json:"status"`
	BookingSource string    `json:"booking_source"`
	NumNights     int       `json:"num_nights"`
}

// CalendarUnit represents a unit row on the calendar.
type CalendarUnit struct {
	UnitID       uuid.UUID       `json:"unit_id"`
	UnitNumber   string          `json:"unit_number"`
	Floor        string          `json:"floor"`
	UnitTypeName string          `json:"unit_type_name"`
	BaseRate     float64         `json:"base_rate"`
	Status       string          `json:"unit_status"`
	Entries      []CalendarEntry `json:"entries"`
}

// CalendarView represents the full calendar grid for a property.
type CalendarView struct {
	PropertyID   uuid.UUID      `json:"property_id"`
	PropertyName string         `json:"property_name"`
	StartDate    time.Time      `json:"start_date"`
	EndDate      time.Time      `json:"end_date"`
	Units        []CalendarUnit `json:"units"`
}

// MoveBookingInput represents a drag/move/extend/shorten action.
type MoveBookingInput struct {
	ReservationID uuid.UUID `json:"reservation_id" validate:"required"`
	NewUnitID     uuid.UUID `json:"new_unit_id" validate:"required"`
	NewCheckIn    string    `json:"new_check_in" validate:"required"`
	NewCheckOut   string    `json:"new_check_out" validate:"required"`
}

// OccupancyStats for a date range.
type OccupancyStats struct {
	Date           time.Time `json:"date"`
	TotalUnits     int       `json:"total_units"`
	OccupiedUnits  int       `json:"occupied_units"`
	AvailableUnits int       `json:"available_units"`
	OccupancyRate  float64   `json:"occupancy_rate"`
}
