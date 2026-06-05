package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/stayflow/stayflow-track/internal/modules/auth/middleware"
	"github.com/stayflow/stayflow-track/internal/modules/calendar/domain"
	"github.com/stayflow/stayflow-track/internal/modules/calendar/service"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
	"github.com/stayflow/stayflow-track/internal/shared/response"
	"github.com/stayflow/stayflow-track/internal/shared/validation"
)

type Handler struct {
	service *service.Service
	log     zerolog.Logger
}

func New(svc *service.Service, log zerolog.Logger) *Handler {
	return &Handler{service: svc, log: log}
}

// GetCalendarView returns the occupancy calendar for a property.
func (h *Handler) GetCalendarView(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	propertyID, err := uuid.Parse(chi.URLParam(r, "propertyID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid property id"))
		return
	}

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if startDate == "" || endDate == "" {
		response.Err(w, apperrors.BadRequest("start_date and end_date are required"))
		return
	}

	view, err := h.service.GetCalendarView(r.Context(), claims.TenantID, propertyID, startDate, endDate)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, view)
}

// MoveBooking handles drag/move/extend/shorten operations.
func (h *Handler) MoveBooking(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input domain.MoveBookingInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	if err := h.service.MoveBooking(r.Context(), claims.TenantID, input); err != nil {
		h.log.Error().Err(err).Msg("move booking failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "booking moved successfully"})
}

// GetOccupancyStats returns occupancy statistics for a property.
func (h *Handler) GetOccupancyStats(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	propertyID, err := uuid.Parse(chi.URLParam(r, "propertyID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid property id"))
		return
	}

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if startDate == "" || endDate == "" {
		response.Err(w, apperrors.BadRequest("start_date and end_date are required"))
		return
	}

	stats, err := h.service.GetOccupancyStats(r.Context(), claims.TenantID, propertyID, startDate, endDate)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, stats)
}
