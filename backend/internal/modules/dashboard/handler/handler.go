package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/stayflow/stayflow-track/internal/modules/auth/middleware"
	"github.com/stayflow/stayflow-track/internal/modules/dashboard/service"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
	"github.com/stayflow/stayflow-track/internal/shared/response"
)

type Handler struct {
	service *service.Service
	log     zerolog.Logger
}

func New(svc *service.Service, log zerolog.Logger) *Handler {
	return &Handler{service: svc, log: log}
}

func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	propertyID, err := uuid.Parse(chi.URLParam(r, "propertyID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid property id"))
		return
	}

	metrics, err := h.service.GetDashboard(r.Context(), claims.TenantID, propertyID)
	if err != nil {
		h.log.Error().Err(err).Msg("get dashboard failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, metrics)
}

func (h *Handler) GetRevenueTrend(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	propertyID, err := uuid.Parse(chi.URLParam(r, "propertyID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid property id"))
		return
	}

	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 90 {
			days = parsed
		}
	}

	// Support start_date/end_date for revenue report date picker
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	if startDate != "" && endDate != "" {
		trends, err := h.service.GetRevenueTrendByRange(r.Context(), claims.TenantID, propertyID, startDate, endDate)
		if err != nil {
			response.Err(w, err)
			return
		}
		response.JSON(w, http.StatusOK, trends)
		return
	}

	trends, err := h.service.GetRevenueTrend(r.Context(), claims.TenantID, propertyID, days)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, trends)
}

// GetDailyCollection returns today's payment breakdown by method.
func (h *Handler) GetDailyCollection(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	propertyID, err := uuid.Parse(chi.URLParam(r, "propertyID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid property id"))
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		date = "today"
	}

	collection, err := h.service.GetDailyCollection(r.Context(), claims.TenantID, propertyID, date)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, collection)
}

// GetOutstandingDues returns all guests with pending balances.
func (h *Handler) GetOutstandingDues(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	propertyID, err := uuid.Parse(chi.URLParam(r, "propertyID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid property id"))
		return
	}

	report, err := h.service.GetOutstandingDues(r.Context(), claims.TenantID, propertyID)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, report)
}

// GetEndOfDaySummary returns the comprehensive night audit report.
func (h *Handler) GetEndOfDaySummary(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	propertyID, err := uuid.Parse(chi.URLParam(r, "propertyID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid property id"))
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		date = "today"
	}

	summary, err := h.service.GetEndOfDaySummary(r.Context(), claims.TenantID, propertyID, date)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, summary)
}

// CloseDay marks the day as audited/closed.
func (h *Handler) CloseDay(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	propertyID, err := uuid.Parse(chi.URLParam(r, "propertyID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid property id"))
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		date = "today"
	}

	if err := h.service.CloseDay(r.Context(), claims.TenantID, propertyID, claims.UserID, date); err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "closed"})
}
