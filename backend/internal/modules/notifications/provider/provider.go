package provider

import (
	"context"

	"github.com/stayflow/stayflow-track/internal/modules/notifications/domain"
)

// Provider is the abstraction for messaging providers (WhatsApp, SMS, Email).
// Implementations can be swapped without changing business logic.
type Provider interface {
	// Name returns the provider identifier (e.g., "twilio", "gupshup", "wati").
	Name() string

	// SendWhatsApp sends a WhatsApp message.
	SendWhatsApp(ctx context.Context, phone string, templateName string, variables map[string]string) (*domain.ProviderResponse, error)

	// SendSMS sends an SMS message.
	SendSMS(ctx context.Context, phone string, message string) (*domain.ProviderResponse, error)

	// SendEmail sends an email.
	SendEmail(ctx context.Context, to string, subject string, body string) (*domain.ProviderResponse, error)

	// GetMessageStatus checks delivery status of a sent message.
	GetMessageStatus(ctx context.Context, messageID string) (string, error)
}
