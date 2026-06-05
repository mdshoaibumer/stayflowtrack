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
		if parsed, err := strconv.Atoi(d); err == nil {
			days = parsed
		}
	}

	trends, err := h.service.GetRevenueTrend(r.Context(), claims.TenantID, propertyID, days)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, trends)
}
