package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type FolioStatus string

const (
	FolioOpen   FolioStatus = "open"
	FolioClosed FolioStatus = "closed"
	FolioVoid   FolioStatus = "void"
)

type Folio struct {
	ID            uuid.UUID       `json:"id"`
	TenantID      uuid.UUID       `json:"tenant_id"`
	ReservationID uuid.UUID       `json:"reservation_id"`
	GuestID       uuid.UUID       `json:"guest_id"`
	PropertyID    uuid.UUID       `json:"property_id"`
	FolioNumber   string          `json:"folio_number"`
	Status        FolioStatus     `json:"status"`
	Subtotal      decimal.Decimal `json:"subtotal"`
	TaxTotal      decimal.Decimal `json:"tax_total"`
	TotalAmount   decimal.Decimal `json:"total_amount"`
	PaidAmount    decimal.Decimal `json:"paid_amount"`
	Balance       decimal.Decimal `json:"balance"`
	Notes         string          `json:"notes,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	ClosedAt      *time.Time      `json:"closed_at,omitempty"`
}

type LineItemCategory string

const (
	CategoryRoomCharge   LineItemCategory = "room_charge"
	CategoryDeposit      LineItemCategory = "deposit"
	CategoryFoodBeverage LineItemCategory = "food_beverage"
	CategoryLaundry      LineItemCategory = "laundry"
	CategoryMinibar      LineItemCategory = "minibar"
	CategoryParking      LineItemCategory = "parking"
	CategorySpa          LineItemCategory = "spa"
	CategoryDamage       LineItemCategory = "damage"
	CategoryLateCheckout LineItemCategory = "late_checkout"
	CategoryExtraBed     LineItemCategory = "extra_bed"
	CategoryOther        LineItemCategory = "other"
)

type LineItem struct {
	ID          uuid.UUID        `json:"id"`
	TenantID    uuid.UUID        `json:"tenant_id"`
	FolioID     uuid.UUID        `json:"folio_id"`
	Category    LineItemCategory `json:"category"`
	Description string           `json:"description"`
	Quantity    int              `json:"quantity"`
	UnitPrice   decimal.Decimal  `json:"unit_price"`
	Amount      decimal.Decimal  `json:"amount"`
	TaxRate     decimal.Decimal  `json:"tax_rate"`
	TaxAmount   decimal.Decimal  `json:"tax_amount"`
	Total       decimal.Decimal  `json:"total"`
	Date        time.Time        `json:"date"`
	IsVoid      bool             `json:"is_void"`
	VoidReason  string           `json:"void_reason,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	CreatedBy   uuid.UUID        `json:"created_by"`
}

type InvoiceStatus string

const (
	InvoiceIssued        InvoiceStatus = "issued"
	InvoicePaid          InvoiceStatus = "paid"
	InvoicePartiallyPaid InvoiceStatus = "partially_paid"
	InvoiceVoid          InvoiceStatus = "void"
	InvoiceRefunded      InvoiceStatus = "refunded"
)

type Invoice struct {
	ID              uuid.UUID         `json:"id"`
	TenantID        uuid.UUID         `json:"tenant_id"`
	FolioID         uuid.UUID         `json:"folio_id"`
	ReservationID   uuid.UUID         `json:"reservation_id"`
	GuestID         uuid.UUID         `json:"guest_id"`
	PropertyID      uuid.UUID         `json:"property_id"`
	InvoiceNumber   string            `json:"invoice_number"`
	Status          InvoiceStatus     `json:"status"`
	GuestName       string            `json:"guest_name"`
	GuestEmail      string            `json:"guest_email,omitempty"`
	GuestPhone      string            `json:"guest_phone,omitempty"`
	GuestAddress    string            `json:"guest_address,omitempty"`
	PropertyName    string            `json:"property_name"`
	PropertyAddress string            `json:"property_address"`
	PropertyGST     string            `json:"property_gst_number,omitempty"`
	Subtotal        decimal.Decimal   `json:"subtotal"`
	CGSTAmount      decimal.Decimal   `json:"cgst_amount"`
	SGSTAmount      decimal.Decimal   `json:"sgst_amount"`
	IGSTAmount      decimal.Decimal   `json:"igst_amount"`
	TotalTax        decimal.Decimal   `json:"total_tax"`
	TotalAmount     decimal.Decimal   `json:"total_amount"`
	PaidAmount      decimal.Decimal   `json:"paid_amount"`
	BalanceDue      decimal.Decimal   `json:"balance_due"`
	CheckInDate     time.Time         `json:"check_in_date"`
	CheckOutDate    time.Time         `json:"check_out_date"`
	NumNights       int               `json:"num_nights"`
	IssuedAt        time.Time         `json:"issued_at"`
	DueDate         *time.Time        `json:"due_date,omitempty"`
	Notes           string            `json:"notes,omitempty"`
	PDFKey          string            `json:"pdf_key,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	LineItems       []InvoiceLineItem `json:"line_items,omitempty"`
}

type InvoiceLineItem struct {
	ID          uuid.UUID       `json:"id"`
	InvoiceID   uuid.UUID       `json:"invoice_id"`
	Category    string          `json:"category"`
	Description string          `json:"description"`
	Quantity    int             `json:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	Amount      decimal.Decimal `json:"amount"`
	TaxRate     decimal.Decimal `json:"tax_rate"`
	TaxAmount   decimal.Decimal `json:"tax_amount"`
	Total       decimal.Decimal `json:"total"`
	Date        time.Time       `json:"date"`
}

type PaymentType string

const (
	PaymentTypePayment        PaymentType = "payment"
	PaymentTypeRefund         PaymentType = "refund"
	PaymentTypeDeposit        PaymentType = "deposit"
	PaymentTypeDepositRelease PaymentType = "deposit_release"
)

type PaymentMethod string

const (
	MethodCash         PaymentMethod = "cash"
	MethodUPI          PaymentMethod = "upi"
	MethodCard         PaymentMethod = "card"
	MethodBankTransfer PaymentMethod = "bank_transfer"
)

type Payment struct {
	ID              uuid.UUID       `json:"id"`
	TenantID        uuid.UUID       `json:"tenant_id"`
	FolioID         uuid.UUID       `json:"folio_id"`
	InvoiceID       *uuid.UUID      `json:"invoice_id,omitempty"`
	PaymentType     PaymentType     `json:"payment_type"`
	PaymentMethod   PaymentMethod   `json:"payment_method"`
	Amount          decimal.Decimal `json:"amount"`
	ReferenceNumber string          `json:"reference_number,omitempty"`
	Notes           string          `json:"notes,omitempty"`
	ReceivedBy      uuid.UUID       `json:"received_by"`
	CreatedAt       time.Time       `json:"created_at"`
}

// GSTBreakdown for invoice generation
type GSTBreakdown struct {
	Rate     decimal.Decimal `json:"rate"`
	Taxable  decimal.Decimal `json:"taxable_amount"`
	CGST     decimal.Decimal `json:"cgst"`
	SGST     decimal.Decimal `json:"sgst"`
	IGST     decimal.Decimal `json:"igst"`
	TotalTax decimal.Decimal `json:"total_tax"`
}
