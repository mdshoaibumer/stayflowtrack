package provider_test

import (
	"context"
	"testing"

	"github.com/stayflow/stayflow-track/internal/modules/notifications/provider"
)

func TestLogProvider_SendWhatsApp(t *testing.T) {
	p := provider.NewLogProvider()

	if p.Name() != "log" {
		t.Errorf("expected name 'log', got %s", p.Name())
	}

	resp, err := p.SendWhatsApp(context.Background(), "+919876543210", "booking_confirm", map[string]string{
		"guest_name": "Test User",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "sent" {
		t.Errorf("expected status 'sent', got %s", resp.Status)
	}
	if resp.MessageID == "" {
		t.Error("expected non-empty message ID")
	}
}

func TestLogProvider_SendSMS(t *testing.T) {
	p := provider.NewLogProvider()

	resp, err := p.SendSMS(context.Background(), "+919876543210", "Your booking is confirmed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "sent" {
		t.Errorf("expected status 'sent', got %s", resp.Status)
	}
}

func TestLogProvider_SendEmail(t *testing.T) {
	p := provider.NewLogProvider()

	resp, err := p.SendEmail(context.Background(), "guest@example.com", "Booking Confirmation", "<html>...</html>")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "sent" {
		t.Errorf("expected status 'sent', got %s", resp.Status)
	}
}

func TestLogProvider_GetMessageStatus(t *testing.T) {
	p := provider.NewLogProvider()

	status, err := p.GetMessageStatus(context.Background(), "test-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "delivered" {
		t.Errorf("expected 'delivered', got %s", status)
	}
}

func TestProviderInterface(t *testing.T) {
	// Verify LogProvider satisfies Provider interface
	var _ provider.Provider = (*provider.LogProvider)(nil)
}
