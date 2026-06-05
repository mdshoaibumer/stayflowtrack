package domain

import (
	"time"

	"github.com/google/uuid"
)

type Channel string

const (
	ChannelWhatsApp Channel = "whatsapp"
	ChannelSMS      Channel = "sms"
	ChannelEmail    Channel = "email"
)

type NotificationStatus string

const (
	StatusPending   NotificationStatus = "pending"
	StatusSent      NotificationStatus = "sent"
	StatusDelivered NotificationStatus = "delivered"
	StatusFailed    NotificationStatus = "failed"
	StatusRead      NotificationStatus = "read"
)

// EventType constants
const (
	EventBookingConfirmation = "booking_confirmation"
	EventCheckInReminder     = "check_in_reminder"
	EventInvoiceDelivery     = "invoice_delivery"
	EventPaymentReminder     = "payment_reminder"
	EventCheckOutReminder    = "check_out_reminder"
	EventPaymentReceived     = "payment_received"
)

type Template struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	EventType    string    `json:"event_type"`
	Channel      Channel   `json:"channel"`
	TemplateName string    `json:"template_name"`
	TemplateBody string    `json:"template_body"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type NotificationLog struct {
	ID                uuid.UUID          `json:"id"`
	TenantID          uuid.UUID          `json:"tenant_id"`
	TemplateID        *uuid.UUID         `json:"template_id,omitempty"`
	EventType         string             `json:"event_type"`
	Channel           Channel            `json:"channel"`
	RecipientPhone    string             `json:"recipient_phone,omitempty"`
	RecipientEmail    string             `json:"recipient_email,omitempty"`
	Status            NotificationStatus `json:"status"`
	Provider          string             `json:"provider,omitempty"`
	ProviderMessageID string             `json:"provider_message_id,omitempty"`
	Payload           map[string]string  `json:"payload,omitempty"`
	ErrorMessage      string             `json:"error_message,omitempty"`
	SentAt            *time.Time         `json:"sent_at,omitempty"`
	DeliveredAt       *time.Time         `json:"delivered_at,omitempty"`
	CreatedAt         time.Time          `json:"created_at"`
}

// SendRequest is what callers provide to send a notification.
type SendRequest struct {
	TenantID       uuid.UUID
	EventType      string
	RecipientPhone string
	RecipientEmail string
	Variables      map[string]string // Template variable substitution
}

// ProviderResponse from the external messaging provider.
type ProviderResponse struct {
	MessageID string
	Status    string
	Error     error
}
