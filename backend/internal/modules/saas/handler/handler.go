package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/stayflow/stayflow-track/internal/modules/auth/middleware"
	"github.com/stayflow/stayflow-track/internal/modules/saas/domain"
	"github.com/stayflow/stayflow-track/internal/modules/saas/razorpay"
	"github.com/stayflow/stayflow-track/internal/modules/saas/service"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
	"github.com/stayflow/stayflow-track/internal/shared/pagination"
	"github.com/stayflow/stayflow-track/internal/shared/response"
	"github.com/stayflow/stayflow-track/internal/shared/validation"
)

type Handler struct {
	service  *service.Service
	razorpay *razorpay.Client
	log      zerolog.Logger
}

func New(svc *service.Service, rp *razorpay.Client, log zerolog.Logger) *Handler {
	return &Handler{service: svc, razorpay: rp, log: log}
}

// Plans

func (h *Handler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.service.ListPlans(r.Context(), true)
	if err != nil {
		response.Err(w, err)
		return
	}
	response.JSON(w, http.StatusOK, plans)
}

// Subscription Management

func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	sub, err := h.service.GetSubscription(r.Context(), claims.TenantID)
	if err != nil {
		response.Err(w, err)
		return
	}
	response.JSON(w, http.StatusOK, sub)
}

func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input domain.CreateSubscriptionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}
	input.TenantID = claims.TenantID

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	sub, err := h.service.CreateSubscription(r.Context(), input)
	if err != nil {
		h.log.Error().Err(err).Msg("create subscription failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, sub)
}

func (h *Handler) ChangePlan(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input domain.ChangePlanInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}
	input.TenantID = claims.TenantID

	sub, err := h.service.ChangePlan(r.Context(), input)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, sub)
}

func (h *Handler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input domain.CancelSubscriptionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}
	input.TenantID = claims.TenantID

	if err := h.service.CancelSubscription(r.Context(), input); err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "subscription cancelled"})
}

// Checkout & Payments

func (h *Handler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input service.CreateCheckoutInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}
	input.TenantID = claims.TenantID

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	checkout, err := h.service.CreateCheckout(r.Context(), input)
	if err != nil {
		h.log.Error().Err(err).Msg("create checkout failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, checkout)
}

func (h *Handler) VerifyPayment(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input service.VerifyPaymentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}
	input.TenantID = claims.TenantID

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	if err := h.service.VerifyPayment(r.Context(), input); err != nil {
		h.log.Error().Err(err).Msg("verify payment failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "payment verified, subscription active"})
}

// Billing History

func (h *Handler) ListBillingEvents(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	pg := pagination.Parse(r)

	events, total, err := h.service.ListBillingEvents(r.Context(), claims.TenantID, pg.PerPage, pg.Offset)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSONWithMeta(w, http.StatusOK, events, &response.Meta{
		Page:       pg.Page,
		PerPage:    pg.PerPage,
		Total:      total,
		TotalPages: pagination.TotalPages(total, pg.PerPage),
	})
}

// Onboarding

func (h *Handler) InitOnboarding(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	propertyID, err := uuid.Parse(chi.URLParam(r, "propertyID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid property id"))
		return
	}

	if err := h.service.InitOnboarding(r.Context(), domain.OnboardPropertyInput{
		TenantID:   claims.TenantID,
		PropertyID: propertyID,
	}); err != nil {
		response.Err(w, err)
		return
	}

	steps, _ := h.service.GetOnboardingSteps(r.Context(), claims.TenantID, propertyID)
	response.JSON(w, http.StatusOK, steps)
}

func (h *Handler) GetOnboardingSteps(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	propertyID, err := uuid.Parse(chi.URLParam(r, "propertyID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid property id"))
		return
	}

	steps, err := h.service.GetOnboardingSteps(r.Context(), claims.TenantID, propertyID)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, steps)
}

func (h *Handler) CompleteStep(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input domain.CompleteStepInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}
	input.TenantID = claims.TenantID

	if err := h.service.CompleteStep(r.Context(), input.TenantID, input.PropertyID, input.StepKey, claims.UserID); err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "step completed"})
}

// Admin APIs (super_admin only)

func (h *Handler) AdminListTenants(w http.ResponseWriter, r *http.Request) {
	pg := pagination.Parse(r)
	status := r.URL.Query().Get("status")

	tenants, total, err := h.service.ListTenants(r.Context(), status, pg.PerPage, pg.Offset)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSONWithMeta(w, http.StatusOK, tenants, &response.Meta{
		Page:       pg.Page,
		PerPage:    pg.PerPage,
		Total:      total,
		TotalPages: pagination.TotalPages(total, pg.PerPage),
	})
}

func (h *Handler) AdminGetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.service.GetSaaSMetrics(r.Context())
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, metrics)
}

func (h *Handler) AdminListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.service.ListPlans(r.Context(), false)
	if err != nil {
		response.Err(w, err)
		return
	}
	response.JSON(w, http.StatusOK, plans)
}

// Razorpay Webhook (public endpoint)
func (h *Handler) RazorpayWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Verify webhook signature — ALWAYS required.
	// If razorpay client is nil (misconfigured), reject all webhooks to fail closed.
	signature := r.Header.Get("X-Razorpay-Signature")
	if h.razorpay == nil {
		h.log.Error().Msg("razorpay webhook received but client not configured — rejecting")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	if signature == "" {
		h.log.Warn().Msg("missing razorpay webhook signature")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !h.razorpay.VerifyWebhookSignature(body, signature) {
		h.log.Warn().Msg("invalid razorpay webhook signature")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var event struct {
		Event   string `json:"event"`
		Payload struct {
			Payment struct {
				Entity struct {
					ID       string `json:"id"`
					Amount   int64  `json:"amount"`
					Currency string `json:"currency"`
					Status   string `json:"status"`
					OrderID  string `json:"order_id"`
				} `json:"entity"`
			} `json:"payment"`
			Subscription struct {
				Entity struct {
					ID         string `json:"id"`
					Status     string `json:"status"`
					CurrentEnd int64  `json:"current_end"`
				} `json:"entity"`
			} `json:"subscription"`
		} `json:"payload"`
	}

	if err := json.Unmarshal(body, &event); err != nil {
		h.log.Error().Err(err).Msg("parse razorpay webhook")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.log.Info().Str("event", event.Event).Msg("razorpay webhook received")

	// Process events asynchronously in production
	switch event.Event {
	case "payment.captured":
		h.log.Info().Str("payment_id", event.Payload.Payment.Entity.ID).Msg("payment captured")
	case "subscription.activated":
		h.log.Info().Str("sub_id", event.Payload.Subscription.Entity.ID).Msg("subscription activated")
	case "subscription.cancelled":
		h.log.Info().Str("sub_id", event.Payload.Subscription.Entity.ID).Msg("subscription cancelled")
	case "payment.failed":
		h.log.Warn().Str("payment_id", event.Payload.Payment.Entity.ID).Msg("payment failed")
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
