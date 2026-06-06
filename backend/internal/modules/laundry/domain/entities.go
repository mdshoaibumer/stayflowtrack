package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type LaundryStatus string

const (
	LaundryReceived  LaundryStatus = "received"
	LaundryWashing   LaundryStatus = "washing"
	LaundryReady     LaundryStatus = "ready"
	LaundryDelivered LaundryStatus = "delivered"
)

type LaundryOrder struct {
	ID            uuid.UUID       `json:"id"`
	TenantID      uuid.UUID       `json:"tenant_id"`
	PropertyID    uuid.UUID       `json:"property_id"`
	ReservationID *uuid.UUID      `json:"reservation_id,omitempty"`
	GuestID       *uuid.UUID      `json:"guest_id,omitempty"`
	FolioID       *uuid.UUID      `json:"folio_id,omitempty"`
	UnitID        *uuid.UUID      `json:"unit_id,omitempty"`
	OrderNumber   string          `json:"order_number"`
	OrderType     string          `json:"order_type"`
	Status        LaundryStatus   `json:"status"`
	TotalItems    int             `json:"total_items"`
	TotalAmount   decimal.Decimal `json:"total_amount"`
	TaxAmount     decimal.Decimal `json:"tax_amount"`
	GrandTotal    decimal.Decimal `json:"grand_total"`
	Notes         string          `json:"notes,omitempty"`
	ReceivedBy    uuid.UUID       `json:"received_by"`
	ReceivedAt    time.Time       `json:"received_at"`
	WashedAt      *time.Time      `json:"washed_at,omitempty"`
	ReadyAt       *time.Time      `json:"ready_at,omitempty"`
	DeliveredAt   *time.Time      `json:"delivered_at,omitempty"`
	DeliveredBy   *uuid.UUID      `json:"delivered_by,omitempty"`
	PostedToFolio bool            `json:"posted_to_folio"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	// Joined
	GuestName  string        `json:"guest_name,omitempty"`
	UnitNumber string        `json:"unit_number,omitempty"`
	Items      []LaundryItem `json:"items,omitempty"`
}

type LaundryItem struct {
	ID          uuid.UUID       `json:"id"`
	OrderID     uuid.UUID       `json:"order_id"`
	ItemType    string          `json:"item_type"`
	Description string          `json:"description,omitempty"`
	Quantity    int             `json:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	Amount      decimal.Decimal `json:"amount"`
	ServiceType string          `json:"service_type"`
	CreatedAt   time.Time       `json:"created_at"`
}

type CreateOrderInput struct {
	PropertyID    uuid.UUID         `json:"property_id" validate:"required"`
	ReservationID *uuid.UUID        `json:"reservation_id"`
	GuestID       *uuid.UUID        `json:"guest_id"`
	UnitID        *uuid.UUID        `json:"unit_id"`
	OrderType     string            `json:"order_type" validate:"required,oneof=guest house"`
	Notes         string            `json:"notes" validate:"omitempty,max=500"`
	Items         []CreateItemInput `json:"items" validate:"required,min=1,dive"`
}

type CreateItemInput struct {
	ItemType    string  `json:"item_type" validate:"required,oneof=bedsheet towel pillow_cover blanket curtain shirt trouser dress saree other"`
	Description string  `json:"description" validate:"omitempty,max=255"`
	Quantity    int     `json:"quantity" validate:"required,min=1,max=100"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"`
	ServiceType string  `json:"service_type" validate:"required,oneof=wash dry_clean iron wash_iron"`
}

type UpdateStatusInput struct {
	OrderID uuid.UUID `json:"order_id" validate:"required"`
	Status  string    `json:"status" validate:"required,oneof=received washing ready delivered"`
}

// LaundryRateCard represents a saved item with default pricing.
type LaundryRateCard struct {
	ID          uuid.UUID       `json:"id"`
	TenantID    uuid.UUID       `json:"tenant_id"`
	PropertyID  uuid.UUID       `json:"property_id"`
	ItemType    string          `json:"item_type"`
	ItemName    string          `json:"item_name"`
	DefaultRate decimal.Decimal `json:"default_rate"`
	ServiceType string          `json:"service_type"`
	IsActive    bool            `json:"is_active"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type CreateRateCardInput struct {
	PropertyID  uuid.UUID `json:"property_id" validate:"required"`
	ItemType    string    `json:"item_type" validate:"required,oneof=bedsheet towel pillow_cover blanket curtain shirt trouser dress saree other"`
	ItemName    string    `json:"item_name" validate:"required,min=1,max=100"`
	DefaultRate float64   `json:"default_rate" validate:"required,gt=0"`
	ServiceType string    `json:"service_type" validate:"required,oneof=wash dry_clean iron wash_iron"`
}

type UpdateRateCardInput struct {
	ID          uuid.UUID `json:"id" validate:"required"`
	ItemName    string    `json:"item_name" validate:"omitempty,min=1,max=100"`
	DefaultRate float64   `json:"default_rate" validate:"omitempty,gt=0"`
	IsActive    *bool     `json:"is_active"`
}
