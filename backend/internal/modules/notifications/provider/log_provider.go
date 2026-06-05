package provider

import (
	"context"
	"fmt"

	"github.com/stayflow/stayflow-track/internal/modules/notifications/domain"
)

// LogProvider is a development/testing provider that logs messages.
type LogProvider struct{}

func NewLogProvider() *LogProvider {
	return &LogProvider{}
}

func (p *LogProvider) Name() string {
	return "log"
}

func (p *LogProvider) SendWhatsApp(_ context.Context, phone string, templateName string, variables map[string]string) (*domain.ProviderResponse, error) {
	fmt.Printf("[LOG-PROVIDER] WhatsApp → %s | template=%s vars=%v\n", phone, templateName, variables)
	return &domain.ProviderResponse{
		MessageID: "log-" + phone,
		Status:    "sent",
	}, nil
}

func (p *LogProvider) SendSMS(_ context.Context, phone string, message string) (*domain.ProviderResponse, error) {
	fmt.Printf("[LOG-PROVIDER] SMS → %s | msg=%s\n", phone, message)
	return &domain.ProviderResponse{
		MessageID: "log-sms-" + phone,
		Status:    "sent",
	}, nil
}

func (p *LogProvider) SendEmail(_ context.Context, to string, subject string, _ string) (*domain.ProviderResponse, error) {
	fmt.Printf("[LOG-PROVIDER] Email → %s | subject=%s\n", to, subject)
	return &domain.ProviderResponse{
		MessageID: "log-email-" + to,
		Status:    "sent",
	}, nil
}

func (p *LogProvider) GetMessageStatus(_ context.Context, messageID string) (string, error) {
	return "delivered", nil
}
