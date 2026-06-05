package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/stayflow/stayflow-track/internal/modules/notifications/domain"
)

// GupshupConfig for Gupshup WhatsApp Business API.
type GupshupConfig struct {
	APIKey  string
	AppName string
	BaseURL string
}

// GupshupProvider implements the Provider interface for Gupshup.
type GupshupProvider struct {
	cfg    GupshupConfig
	client *http.Client
}

func NewGupshupProvider(cfg GupshupConfig) *GupshupProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.gupshup.io/wa/api/v1"
	}
	return &GupshupProvider{
		cfg: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (p *GupshupProvider) Name() string {
	return "gupshup"
}

func (p *GupshupProvider) SendWhatsApp(ctx context.Context, phone string, templateName string, variables map[string]string) (*domain.ProviderResponse, error) {
	payload := map[string]interface{}{
		"channel":     "whatsapp",
		"source":      p.cfg.AppName,
		"destination": phone,
		"message": map[string]interface{}{
			"type":     "template",
			"template": templateName,
			"params":   variables,
		},
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", p.cfg.BaseURL+"/msg", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", p.cfg.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		Status    string `json:"status"`
		MessageID string `json:"messageId"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&result)

	if resp.StatusCode >= 400 {
		return &domain.ProviderResponse{Error: fmt.Errorf("gupshup error: status %d", resp.StatusCode)}, nil
	}

	return &domain.ProviderResponse{
		MessageID: result.MessageID,
		Status:    "sent",
	}, nil
}

func (p *GupshupProvider) SendSMS(ctx context.Context, phone string, message string) (*domain.ProviderResponse, error) {
	// Gupshup SMS API
	return &domain.ProviderResponse{Status: "unsupported"}, fmt.Errorf("SMS not implemented for gupshup")
}

func (p *GupshupProvider) SendEmail(_ context.Context, _ string, _ string, _ string) (*domain.ProviderResponse, error) {
	return nil, fmt.Errorf("email not supported by gupshup provider")
}

func (p *GupshupProvider) GetMessageStatus(ctx context.Context, messageID string) (string, error) {
	// Would call Gupshup status API
	return "sent", nil
}
