package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CheckInDetails struct {
	ID                uuid.UUID       `json:"id"`
	TenantID          uuid.UUID       `json:"tenant_id"`
	ReservationID     uuid.UUID       `json:"reservation_id"`
	AssignedUnitID    uuid.UUID       `json:"assigned_unit_id"`
	DepositAmount     decimal.Decimal `json:"deposit_amount"`
	DepositMethod     string          `json:"deposit_method,omitempty"`
	DepositReference  string          `json:"deposit_reference,omitempty"`
	IDDocumentType    string          `json:"id_document_type,omitempty"`
	IDDocumentKey     string          `json:"id_document_key,omitempty"`
	GuestSignatureKey string          `json:"guest_signature_key,omitempty"`
	CheckedInBy       uuid.UUID       `json:"checked_in_by"`
	Notes             string          `json:"notes,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
}

type CheckInInput struct {
	ReservationID    uuid.UUID       `json:"reservation_id" validate:"required"`
	AssignedUnitID   uuid.UUID       `json:"assigned_unit_id" validate:"required"`
	DepositAmount    decimal.Decimal `json:"deposit_amount" validate:"required"`
	DepositMethod    string          `json:"deposit_method" validate:"required,oneof=cash upi card bank_transfer cheque"`
	DepositReference string          `json:"deposit_reference" validate:"omitempty,max=255"`
	IDDocumentType   string          `json:"id_document_type" validate:"required,oneof=aadhaar passport driving_license voter_id pan_card"`
	IDDocumentNumber string          `json:"id_document_number" validate:"required,min=4,max=50"`
	Notes            string          `json:"notes" validate:"omitempty,max=1000"`
}

type CheckOutInput struct {
	ReservationID      uuid.UUID       `json:"reservation_id" validate:"required"`
	LateCheckoutCharge decimal.Decimal `json:"late_checkout_charge"`
	Notes              string          `json:"notes" validate:"omitempty,max=1000"`
}

type CheckOutResult struct {
	ReservationID  uuid.UUID       `json:"reservation_id"`
	GuestName      string          `json:"guest_name"`
	UnitNumber     string          `json:"unit_number"`
	CheckInDate    time.Time       `json:"check_in_date"`
	CheckOutDate   time.Time       `json:"check_out_date"`
	ActualCheckOut time.Time       `json:"actual_check_out"`
	TotalCharges   decimal.Decimal `json:"total_charges"`
	TotalPayments  decimal.Decimal `json:"total_payments"`
	Balance        decimal.Decimal `json:"balance"`
	InvoiceID      uuid.UUID       `json:"invoice_id,omitempty"`
	IsLateCheckOut bool            `json:"is_late_check_out"`
	LateByHours    float64         `json:"late_by_hours,omitempty"`
}

// WalkInInput combines guest creation + reservation + check-in in one step.
type WalkInInput struct {
	PropertyID uuid.UUID `json:"property_id" validate:"required"`
	UnitID     uuid.UUID `json:"unit_id" validate:"required"`

	// Guest details (created inline)
	GuestFirstName string `json:"guest_first_name" validate:"required,min=1,max=100"`
	GuestLastName  string `json:"guest_last_name" validate:"required,min=1,max=100"`
	GuestPhone     string `json:"guest_phone" validate:"required,min=10,max=15"`
	GuestEmail     string `json:"guest_email" validate:"omitempty,email"`

	// Stay details
	CheckOutDate string          `json:"check_out_date" validate:"required"`
	NumGuests    int             `json:"num_guests" validate:"required,min=1,max=20"`
	RatePerNight decimal.Decimal `json:"rate_per_night" validate:"required"`

	// Check-in details
	DepositAmount    decimal.Decimal `json:"deposit_amount" validate:"required"`
	DepositMethod    string          `json:"deposit_method" validate:"required,oneof=cash upi card bank_transfer cheque"`
	DepositReference string          `json:"deposit_reference" validate:"omitempty,max=255"`
	IDDocumentType   string          `json:"id_document_type" validate:"required,oneof=aadhaar passport driving_license voter_id pan_card"`
	IDDocumentNumber string          `json:"id_document_number" validate:"required,min=4,max=50"`
	Notes            string          `json:"notes" validate:"omitempty,max=1000"`
}

type WalkInResult struct {
	GuestID       uuid.UUID `json:"guest_id"`
	ReservationID uuid.UUID `json:"reservation_id"`
	FolioID       uuid.UUID `json:"folio_id"`
	GuestName     string    `json:"guest_name"`
	UnitNumber    string    `json:"unit_number"`
	CheckInDate   string    `json:"check_in_date"`
	CheckOutDate  string    `json:"check_out_date"`
	TotalAmount   string    `json:"total_amount"`
}

// FolioSummary is returned before checkout so owner can see bill.
type FolioSummary struct {
	FolioID        uuid.UUID       `json:"folio_id"`
	ReservationID  uuid.UUID       `json:"reservation_id"`
	GuestName      string          `json:"guest_name"`
	UnitNumber     string          `json:"unit_number"`
	CheckInDate    time.Time       `json:"check_in_date"`
	CheckOutDate   time.Time       `json:"check_out_date"`
	Nights         int             `json:"nights"`
	Subtotal       decimal.Decimal `json:"subtotal"`
	TaxTotal       decimal.Decimal `json:"tax_total"`
	TotalAmount    decimal.Decimal `json:"total_amount"`
	PaidAmount     decimal.Decimal `json:"paid_amount"`
	Balance        decimal.Decimal `json:"balance"`
	DepositHeld    decimal.Decimal `json:"deposit_held"`
	IsLateCheckOut bool            `json:"is_late_check_out"`
	LateByHours    float64         `json:"late_by_hours"`
	LineItems      []FolioLineItem `json:"line_items"`
	Payments       []FolioPayment  `json:"payments"`
}

type FolioLineItem struct {
	Category    string          `json:"category"`
	Description string          `json:"description"`
	Quantity    int             `json:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	Amount      decimal.Decimal `json:"amount"`
	TaxAmount   decimal.Decimal `json:"tax_amount"`
	Total       decimal.Decimal `json:"total"`
}

type FolioPayment struct {
	PaymentType   string          `json:"payment_type"`
	PaymentMethod string          `json:"payment_method"`
	Amount        decimal.Decimal `json:"amount"`
	Reference     string          `json:"reference,omitempty"`
	PaidAt        time.Time       `json:"paid_at"`
}
