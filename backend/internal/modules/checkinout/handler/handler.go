package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/stayflow/stayflow-track/internal/modules/auth/middleware"
	"github.com/stayflow/stayflow-track/internal/modules/checkinout/domain"
	"github.com/stayflow/stayflow-track/internal/modules/checkinout/service"
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

func (h *Handler) CheckIn(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input domain.CheckInInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	result, err := h.service.CheckIn(r.Context(), claims.TenantID, claims.UserID, input)
	if err != nil {
		h.log.Error().Err(err).Str("reservation_id", input.ReservationID.String()).Msg("check-in failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) CheckOut(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input domain.CheckOutInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	result, err := h.service.CheckOut(r.Context(), claims.TenantID, claims.UserID, input)
	if err != nil {
		h.log.Error().Err(err).Str("reservation_id", input.ReservationID.String()).Msg("check-out failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}
