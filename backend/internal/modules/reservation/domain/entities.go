package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ReservationStatus string

const (
	StatusPending    ReservationStatus = "pending"
	StatusConfirmed  ReservationStatus = "confirmed"
	StatusCheckedIn  ReservationStatus = "checked_in"
	StatusCheckedOut ReservationStatus = "checked_out"
	StatusCancelled  ReservationStatus = "cancelled"
)

func (s ReservationStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusConfirmed, StatusCheckedIn, StatusCheckedOut, StatusCancelled:
		return true
	}
	return false
}

func (s ReservationStatus) CanTransitionTo(next ReservationStatus) bool {
	transitions := map[ReservationStatus][]ReservationStatus{
		StatusPending:   {StatusConfirmed, StatusCancelled},
		StatusConfirmed: {StatusCheckedIn, StatusCancelled},
		StatusCheckedIn: {StatusCheckedOut},
	}

	allowed, exists := transitions[s]
	if !exists {
		return false
	}

	for _, a := range allowed {
		if a == next {
			return true
		}
	}
	return false
}

type BookingSource string

const (
	SourceWalkIn     BookingSource = "walk_in"
	SourcePhone      BookingSource = "phone"
	SourceWhatsApp   BookingSource = "whatsapp"
	SourceBookingCom BookingSource = "booking_com"
	SourceAirbnb     BookingSource = "airbnb"
	SourceOther      BookingSource = "other"
)

func (s BookingSource) IsValid() bool {
	switch s {
	case SourceWalkIn, SourcePhone, SourceWhatsApp, SourceBookingCom, SourceAirbnb, SourceOther:
		return true
	}
	return false
}

type Reservation struct {
	ID                 uuid.UUID         `json:"id"`
	TenantID           uuid.UUID         `json:"tenant_id"`
	PropertyID         uuid.UUID         `json:"property_id"`
	UnitID             uuid.UUID         `json:"unit_id"`
	GuestID            uuid.UUID         `json:"guest_id"`
	BookingSource      BookingSource     `json:"booking_source"`
	Status             ReservationStatus `json:"status"`
	CheckInDate        time.Time         `json:"check_in_date"`
	CheckOutDate       time.Time         `json:"check_out_date"`
	ActualCheckIn      *time.Time        `json:"actual_check_in,omitempty"`
	ActualCheckOut     *time.Time        `json:"actual_check_out,omitempty"`
	NumGuests          int               `json:"num_guests"`
	RatePerNight       decimal.Decimal   `json:"rate_per_night"`
	TotalAmount        decimal.Decimal   `json:"total_amount"`
	AdvanceAmount      decimal.Decimal   `json:"advance_amount,omitempty"`
	AdvanceMethod      string            `json:"advance_method,omitempty"`
	AdvanceReference   string            `json:"advance_reference,omitempty"`
	Notes              string            `json:"notes,omitempty"`
	CancellationReason string            `json:"cancellation_reason,omitempty"`
	CancelledAt        *time.Time        `json:"cancelled_at,omitempty"`
	ExternalBookingID  string            `json:"external_booking_id,omitempty"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`

	// Joined fields
	GuestFirstName string `json:"guest_first_name,omitempty"`
	GuestLastName  string `json:"guest_last_name,omitempty"`
	GuestPhone     string `json:"guest_phone,omitempty"`
	UnitNumber     string `json:"unit_number,omitempty"`
	UnitTypeName   string `json:"unit_type_name,omitempty"`
	PropertyName   string `json:"property_name,omitempty"`
}

type AvailableUnit struct {
	ID           uuid.UUID       `json:"id"`
	UnitNumber   string          `json:"unit_number"`
	Floor        string          `json:"floor,omitempty"`
	UnitTypeName string          `json:"unit_type_name"`
	BaseRate     decimal.Decimal `json:"base_rate"`
}
