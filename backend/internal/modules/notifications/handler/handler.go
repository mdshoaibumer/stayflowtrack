package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/rs/zerolog"

	"github.com/stayflow/stayflow-track/internal/modules/auth/middleware"
	"github.com/stayflow/stayflow-track/internal/modules/notifications/domain"
	"github.com/stayflow/stayflow-track/internal/modules/notifications/service"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
	"github.com/stayflow/stayflow-track/internal/shared/pagination"
	"github.com/stayflow/stayflow-track/internal/shared/response"
)

type Handler struct {
	service *service.Service
	log     zerolog.Logger
}

func New(svc *service.Service, log zerolog.Logger) *Handler {
	return &Handler{service: svc, log: log}
}

type UpsertTemplateInput struct {
	EventType    string `json:"event_type" validate:"required"`
	Channel      string `json:"channel" validate:"required,oneof=whatsapp sms email"`
	TemplateName string `json:"template_name" validate:"required,max=100"`
	TemplateBody string `json:"template_body" validate:"required"`
	IsActive     bool   `json:"is_active"`
}

type SendTestInput struct {
	EventType      string            `json:"event_type" validate:"required"`
	RecipientPhone string            `json:"recipient_phone" validate:"omitempty,max=20"`
	RecipientEmail string            `json:"recipient_email" validate:"omitempty,email"`
	Variables      map[string]string `json:"variables"`
}

func (h *Handler) UpsertTemplate(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input UpsertTemplateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	t := &domain.Template{
		TenantID:     claims.TenantID,
		EventType:    input.EventType,
		Channel:      domain.Channel(input.Channel),
		TemplateName: input.TemplateName,
		TemplateBody: input.TemplateBody,
		IsActive:     input.IsActive,
	}

	if err := h.service.UpsertTemplate(r.Context(), t); err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, t)
}

func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	templates, err := h.service.ListTemplates(r.Context(), claims.TenantID)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, templates)
}

func (h *Handler) SendTest(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input SendTestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if input.RecipientPhone == "" && input.RecipientEmail == "" {
		response.Err(w, apperrors.BadRequest("recipient_phone or recipient_email required"))
		return
	}

	err := h.service.Send(r.Context(), domain.SendRequest{
		TenantID:       claims.TenantID,
		EventType:      input.EventType,
		RecipientPhone: input.RecipientPhone,
		RecipientEmail: input.RecipientEmail,
		Variables:      input.Variables,
	})
	if err != nil {
		response.Err(w, apperrors.Internal(err))
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "notification sent"})
}

func (h *Handler) ListLogs(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	pg := pagination.Parse(r)

	logs, total, err := h.service.ListLogs(r.Context(), claims.TenantID, pg.PerPage, pg.Offset)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSONWithMeta(w, http.StatusOK, logs, &response.Meta{
		Page:       pg.Page,
		PerPage:    pg.PerPage,
		Total:      total,
		TotalPages: pagination.TotalPages(total, pg.PerPage),
	})
}

// Webhook handles delivery status callbacks from providers.
func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	// Verify webhook signature (HMAC-SHA256)
	webhookSecret := os.Getenv("NOTIFICATION_WEBHOOK_SECRET")
	if webhookSecret == "" {
		h.log.Warn().Msg("NOTIFICATION_WEBHOOK_SECRET not configured, rejecting webhook")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 64*1024)) // Max 64KB webhook payload
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	signature := r.Header.Get("X-Webhook-Signature")
	if signature == "" {
		signature = r.Header.Get("X-Hub-Signature-256")
	}

	if !verifyWebhookSignature(body, signature, webhookSecret) {
		h.log.Warn().Str("ip", r.RemoteAddr).Msg("webhook signature verification failed")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Process the webhook payload
	h.log.Info().Str("provider", r.Header.Get("X-Provider")).Msg("notification webhook received")

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// verifyWebhookSignature checks HMAC-SHA256 signature.
func verifyWebhookSignature(payload []byte, signature, secret string) bool {
	if signature == "" || secret == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// Strip "sha256=" prefix if present
	if len(signature) > 7 && signature[:7] == "sha256=" {
		signature = signature[7:]
	}

	return hmac.Equal([]byte(expectedSig), []byte(signature))
}
