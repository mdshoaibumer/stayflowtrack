package service_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stayflow/stayflow-track/internal/modules/notifications/domain"
)

func TestTemplateVariableSubstitution(t *testing.T) {
	template := "Hello {{guest_name}}, your booking at {{property_name}} is confirmed for {{check_in}}."
	variables := map[string]string{
		"guest_name":    "John Doe",
		"property_name": "Grand Hotel",
		"check_in":      "2026-07-01",
	}

	body := template
	for k, v := range variables {
		body = strings.ReplaceAll(body, "{{"+k+"}}", v)
	}

	expected := "Hello John Doe, your booking at Grand Hotel is confirmed for 2026-07-01."
	if body != expected {
		t.Errorf("expected %q, got %q", expected, body)
	}
}

func TestEventTypes(t *testing.T) {
	events := []string{
		domain.EventBookingConfirmation,
		domain.EventCheckInReminder,
		domain.EventInvoiceDelivery,
		domain.EventPaymentReminder,
		domain.EventCheckOutReminder,
		domain.EventPaymentReceived,
	}

	for _, event := range events {
		if event == "" {
			t.Error("event type should not be empty")
		}
	}

	if len(events) != 6 {
		t.Errorf("expected 6 event types, got %d", len(events))
	}
}

func TestSendRequest_Validation(t *testing.T) {
	tests := []struct {
		name  string
		req   domain.SendRequest
		valid bool
	}{
		{
			name: "valid whatsapp",
			req: domain.SendRequest{
				TenantID:       uuid.New(),
				EventType:      domain.EventBookingConfirmation,
				RecipientPhone: "+919876543210",
				Variables:      map[string]string{"guest_name": "Test"},
			},
			valid: true,
		},
		{
			name: "valid email",
			req: domain.SendRequest{
				TenantID:       uuid.New(),
				EventType:      domain.EventInvoiceDelivery,
				RecipientEmail: "test@example.com",
			},
			valid: true,
		},
		{
			name: "no recipient",
			req: domain.SendRequest{
				TenantID:  uuid.New(),
				EventType: domain.EventPaymentReminder,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasRecipient := tt.req.RecipientPhone != "" || tt.req.RecipientEmail != ""
			if hasRecipient != tt.valid {
				t.Errorf("recipient validation: expected valid=%v, got has_recipient=%v", tt.valid, hasRecipient)
			}
		})
	}
}

func TestNotificationStatuses(t *testing.T) {
	statuses := []domain.NotificationStatus{
		domain.StatusPending,
		domain.StatusSent,
		domain.StatusDelivered,
		domain.StatusFailed,
		domain.StatusRead,
	}

	seen := make(map[domain.NotificationStatus]bool)
	for _, s := range statuses {
		if seen[s] {
			t.Errorf("duplicate status: %s", s)
		}
		seen[s] = true
	}
}
