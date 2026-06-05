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
	DepositAmount    decimal.Decimal `json:"deposit_amount"`
	DepositMethod    string          `json:"deposit_method" validate:"omitempty,oneof=cash upi card bank_transfer"`
	DepositReference string          `json:"deposit_reference" validate:"omitempty,max=255"`
	IDDocumentType   string          `json:"id_document_type" validate:"omitempty,oneof=aadhaar passport driving_license voter_id"`
	Notes            string          `json:"notes" validate:"omitempty,max=1000"`
}

type CheckOutInput struct {
	ReservationID uuid.UUID `json:"reservation_id" validate:"required"`
	Notes         string    `json:"notes" validate:"omitempty,max=1000"`
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
}
