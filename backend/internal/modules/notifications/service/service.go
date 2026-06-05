package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stayflow/stayflow-track/internal/modules/notifications/domain"
	"github.com/stayflow/stayflow-track/internal/modules/notifications/provider"
	"github.com/stayflow/stayflow-track/internal/modules/notifications/repository"
)

type Service struct {
	repo     *repository.Repository
	provider provider.Provider
	log      zerolog.Logger
}

func New(repo *repository.Repository, prov provider.Provider, log zerolog.Logger) *Service {
	return &Service{repo: repo, provider: prov, log: log}
}

// Send dispatches a notification via the configured provider.
func (s *Service) Send(ctx context.Context, req domain.SendRequest) error {
	// Look up template
	channel := domain.ChannelWhatsApp
	if req.RecipientPhone == "" && req.RecipientEmail != "" {
		channel = domain.ChannelEmail
	}

	template, err := s.repo.GetTemplate(ctx, req.TenantID, req.EventType, channel)
	if err != nil {
		s.log.Warn().Str("event", req.EventType).Msg("no template found, skipping notification")
		return nil // Don't fail operations if template missing
	}

	// Resolve template body with variables
	body := template.TemplateBody
	for k, v := range req.Variables {
		body = strings.ReplaceAll(body, "{{"+k+"}}", v)
	}

	// Send via provider
	var resp *domain.ProviderResponse
	switch channel {
	case domain.ChannelWhatsApp:
		resp, err = s.provider.SendWhatsApp(ctx, req.RecipientPhone, template.TemplateName, req.Variables)
	case domain.ChannelSMS:
		resp, err = s.provider.SendSMS(ctx, req.RecipientPhone, body)
	case domain.ChannelEmail:
		resp, err = s.provider.SendEmail(ctx, req.RecipientEmail, req.EventType, body)
	}

	// Log the notification
	now := time.Now()
	logEntry := &domain.NotificationLog{
		TenantID:       req.TenantID,
		TemplateID:     &template.ID,
		EventType:      req.EventType,
		Channel:        channel,
		RecipientPhone: req.RecipientPhone,
		RecipientEmail: req.RecipientEmail,
		Provider:       s.provider.Name(),
		Payload:        req.Variables,
	}

	if err != nil || (resp != nil && resp.Error != nil) {
		logEntry.Status = domain.StatusFailed
		if err != nil {
			logEntry.ErrorMessage = err.Error()
		} else {
			logEntry.ErrorMessage = resp.Error.Error()
		}
	} else {
		logEntry.Status = domain.StatusSent
		logEntry.SentAt = &now
		if resp != nil {
			logEntry.ProviderMessageID = resp.MessageID
		}
	}

	if logErr := s.repo.CreateLog(ctx, logEntry); logErr != nil {
		s.log.Error().Err(logErr).Msg("failed to log notification")
	}

	return err
}

// SendBookingConfirmation sends a booking confirmation notification.
func (s *Service) SendBookingConfirmation(ctx context.Context, tenantID uuid.UUID, phone, guestName, propertyName, checkIn, checkOut, confirmationNo string) {
	s.Send(ctx, domain.SendRequest{
		TenantID:       tenantID,
		EventType:      domain.EventBookingConfirmation,
		RecipientPhone: phone,
		Variables: map[string]string{
			"guest_name":      guestName,
			"property_name":   propertyName,
			"check_in":        checkIn,
			"check_out":       checkOut,
			"confirmation_no": confirmationNo,
		},
	})
}

// SendCheckInReminder sends a check-in reminder (typically 1 day before).
func (s *Service) SendCheckInReminder(ctx context.Context, tenantID uuid.UUID, phone, guestName, propertyName, checkIn string) {
	s.Send(ctx, domain.SendRequest{
		TenantID:       tenantID,
		EventType:      domain.EventCheckInReminder,
		RecipientPhone: phone,
		Variables: map[string]string{
			"guest_name":    guestName,
			"property_name": propertyName,
			"check_in":      checkIn,
		},
	})
}

// SendInvoice sends the invoice via WhatsApp.
func (s *Service) SendInvoice(ctx context.Context, tenantID uuid.UUID, phone, guestName, invoiceNo, amount, pdfURL string) {
	s.Send(ctx, domain.SendRequest{
		TenantID:       tenantID,
		EventType:      domain.EventInvoiceDelivery,
		RecipientPhone: phone,
		Variables: map[string]string{
			"guest_name": guestName,
			"invoice_no": invoiceNo,
			"amount":     amount,
			"pdf_url":    pdfURL,
		},
	})
}

// SendPaymentReminder sends a payment reminder.
func (s *Service) SendPaymentReminder(ctx context.Context, tenantID uuid.UUID, phone, guestName, amount, dueDate string) {
	s.Send(ctx, domain.SendRequest{
		TenantID:       tenantID,
		EventType:      domain.EventPaymentReminder,
		RecipientPhone: phone,
		Variables: map[string]string{
			"guest_name": guestName,
			"amount":     amount,
			"due_date":   dueDate,
		},
	})
}

// Template management
func (s *Service) UpsertTemplate(ctx context.Context, t *domain.Template) error {
	return s.repo.UpsertTemplate(ctx, t)
}

func (s *Service) ListTemplates(ctx context.Context, tenantID uuid.UUID) ([]domain.Template, error) {
	return s.repo.ListTemplates(ctx, tenantID)
}

func (s *Service) ListLogs(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.NotificationLog, int64, error) {
	return s.repo.ListLogs(ctx, tenantID, limit, offset)
}
