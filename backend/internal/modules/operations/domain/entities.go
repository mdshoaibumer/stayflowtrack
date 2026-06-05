package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// MaintenanceBlock represents a period where a unit is out of service.
type MaintenanceBlock struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	PropertyID uuid.UUID `json:"property_id"`
	UnitID     uuid.UUID `json:"unit_id"`
	Reason     string    `json:"reason"`
	BlockType  string    `json:"block_type"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	CreatedBy  uuid.UUID `json:"created_by"`
	Notes      string    `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// StayExtension records a guest extending their stay.
type StayExtension struct {
	ID               uuid.UUID       `json:"id"`
	TenantID         uuid.UUID       `json:"tenant_id"`
	ReservationID    uuid.UUID       `json:"reservation_id"`
	OriginalCheckOut time.Time       `json:"original_check_out"`
	NewCheckOut      time.Time       `json:"new_check_out"`
	AdditionalNights int             `json:"additional_nights"`
	RatePerNight     decimal.Decimal `json:"rate_per_night"`
	AdditionalAmount decimal.Decimal `json:"additional_amount"`
	Reason           string          `json:"reason,omitempty"`
	ExtendedBy       uuid.UUID       `json:"extended_by"`
	CreatedAt        time.Time       `json:"created_at"`
}

// RoomMove records a guest being moved between units.
type RoomMove struct {
	ID            uuid.UUID       `json:"id"`
	TenantID      uuid.UUID       `json:"tenant_id"`
	ReservationID uuid.UUID       `json:"reservation_id"`
	FromUnitID    uuid.UUID       `json:"from_unit_id"`
	ToUnitID      uuid.UUID       `json:"to_unit_id"`
	Reason        string          `json:"reason"`
	RateChange    decimal.Decimal `json:"rate_change"`
	MovedBy       uuid.UUID       `json:"moved_by"`
	MovedAt       time.Time       `json:"moved_at"`
}

// CorporateAccount represents a corporate billing agreement.
type CorporateAccount struct {
	ID                     uuid.UUID       `json:"id"`
	TenantID               uuid.UUID       `json:"tenant_id"`
	CompanyName            string          `json:"company_name"`
	ContactPerson          string          `json:"contact_person,omitempty"`
	Email                  string          `json:"email,omitempty"`
	Phone                  string          `json:"phone,omitempty"`
	GSTNumber              string          `json:"gst_number,omitempty"`
	BillingAddress         string          `json:"billing_address,omitempty"`
	CreditLimit            decimal.Decimal `json:"credit_limit"`
	PaymentTermsDays       int             `json:"payment_terms_days"`
	NegotiatedRateDiscount decimal.Decimal `json:"negotiated_rate_discount"`
	IsActive               bool            `json:"is_active"`
	Notes                  string          `json:"notes,omitempty"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
}

// Deposit represents a security deposit collected during check-in.
type Deposit struct {
	ID              uuid.UUID       `json:"id"`
	TenantID        uuid.UUID       `json:"tenant_id"`
	ReservationID   uuid.UUID       `json:"reservation_id"`
	FolioID         *uuid.UUID      `json:"folio_id,omitempty"`
	Amount          decimal.Decimal `json:"amount"`
	Method          string          `json:"method"`
	ReferenceNumber string          `json:"reference_number,omitempty"`
	Status          DepositStatus   `json:"status"`
	CollectedAt     time.Time       `json:"collected_at"`
	CollectedBy     uuid.UUID       `json:"collected_by"`
	ReleasedAt      *time.Time      `json:"released_at,omitempty"`
	ReleasedBy      *uuid.UUID      `json:"released_by,omitempty"`
	RefundAmount    decimal.Decimal `json:"refund_amount,omitempty"`
	RefundMethod    string          `json:"refund_method,omitempty"`
	RefundReference string          `json:"refund_reference,omitempty"`
	Notes           string          `json:"notes,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type DepositStatus string

const (
	DepositHeld      DepositStatus = "held"
	DepositApplied   DepositStatus = "applied"
	DepositRefunded  DepositStatus = "refunded"
	DepositForfeited DepositStatus = "forfeited"
)

// NoShowInput for marking a reservation as no-show.
type NoShowInput struct {
	ReservationID uuid.UUID       `json:"reservation_id" validate:"required"`
	ChargeAmount  decimal.Decimal `json:"charge_amount"`
}

// ExtendStayInput for extending a guest's stay.
type ExtendStayInput struct {
	ReservationID uuid.UUID       `json:"reservation_id" validate:"required"`
	NewCheckOut   string          `json:"new_check_out" validate:"required"`
	RatePerNight  decimal.Decimal `json:"rate_per_night"`
	Reason        string          `json:"reason" validate:"omitempty,max=500"`
}

// RoomMoveInput for moving a guest to a different room.
type RoomMoveInput struct {
	ReservationID uuid.UUID       `json:"reservation_id" validate:"required"`
	ToUnitID      uuid.UUID       `json:"to_unit_id" validate:"required"`
	Reason        string          `json:"reason" validate:"required,max=500"`
	RateChange    decimal.Decimal `json:"rate_change"`
}

// MaintenanceBlockInput for creating a maintenance block.
type MaintenanceBlockInput struct {
	PropertyID uuid.UUID `json:"property_id" validate:"required"`
	UnitID     uuid.UUID `json:"unit_id" validate:"required"`
	Reason     string    `json:"reason" validate:"required,max=500"`
	BlockType  string    `json:"block_type" validate:"required,oneof=maintenance renovation deep_cleaning owner_use other"`
	StartDate  string    `json:"start_date" validate:"required"`
	EndDate    string    `json:"end_date" validate:"required"`
	Notes      string    `json:"notes" validate:"omitempty,max=1000"`
}

// RefundDepositInput for processing deposit refunds at check-out.
type RefundDepositInput struct {
	DepositID       uuid.UUID       `json:"deposit_id" validate:"required"`
	RefundAmount    decimal.Decimal `json:"refund_amount" validate:"required"`
	RefundMethod    string          `json:"refund_method" validate:"required,oneof=cash upi card bank_transfer"`
	RefundReference string          `json:"refund_reference" validate:"omitempty,max=255"`
	Notes           string          `json:"notes" validate:"omitempty,max=500"`
}
