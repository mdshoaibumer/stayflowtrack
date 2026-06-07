package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/stayflow/stayflow-track/internal/modules/auth/middleware"
	resrepo "github.com/stayflow/stayflow-track/internal/modules/reservation/repository"
	"github.com/stayflow/stayflow-track/internal/modules/reservation/service"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
	"github.com/stayflow/stayflow-track/internal/shared/pagination"
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

func (h *Handler) CreateReservation(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input service.CreateReservationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	res, err := h.service.CreateReservation(r.Context(), claims.TenantID, input)
	if err != nil {
		h.log.Error().Err(err).Msg("create reservation failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, res)
}

func (h *Handler) GetReservation(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	resID, err := uuid.Parse(chi.URLParam(r, "reservationID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid reservation id"))
		return
	}

	res, err := h.service.GetReservation(r.Context(), resID, claims.TenantID)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, res)
}

func (h *Handler) ListReservations(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	pg := pagination.Parse(r)

	params := resrepo.ListParams{
		TenantID: claims.TenantID,
		Status:   r.URL.Query().Get("status"),
		Limit:    pg.PerPage,
		Offset:   pg.Offset,
	}

	if propID := r.URL.Query().Get("property_id"); propID != "" {
		id, err := uuid.Parse(propID)
		if err == nil {
			params.PropertyID = &id
		}
	}

	reservations, total, err := h.service.ListReservations(r.Context(), params)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSONWithMeta(w, http.StatusOK, reservations, &response.Meta{
		Page:       pg.Page,
		PerPage:    pg.PerPage,
		Total:      total,
		TotalPages: pagination.TotalPages(total, pg.PerPage),
	})
}

func (h *Handler) UpdateReservation(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	resID, err := uuid.Parse(chi.URLParam(r, "reservationID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid reservation id"))
		return
	}

	var input service.UpdateReservationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	res, err := h.service.UpdateReservation(r.Context(), resID, claims.TenantID, input)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, res)
}

func (h *Handler) CancelReservation(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	resID, err := uuid.Parse(chi.URLParam(r, "reservationID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid reservation id"))
		return
	}

	var input service.CancelReservationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	if err := h.service.CancelReservation(r.Context(), resID, claims.TenantID, input); err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "reservation cancelled"})
}

func (h *Handler) ConfirmReservation(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	resID, err := uuid.Parse(chi.URLParam(r, "reservationID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid reservation id"))
		return
	}

	if err := h.service.ConfirmReservation(r.Context(), resID, claims.TenantID); err != nil {
		h.log.Error().Err(err).Msg("confirm reservation failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "reservation confirmed"})
}

func (h *Handler) CheckIn(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	resID, err := uuid.Parse(chi.URLParam(r, "reservationID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid reservation id"))
		return
	}

	if err := h.service.CheckIn(r.Context(), resID, claims.TenantID); err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "checked in successfully"})
}

func (h *Handler) CheckOut(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	resID, err := uuid.Parse(chi.URLParam(r, "reservationID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid reservation id"))
		return
	}

	if err := h.service.CheckOut(r.Context(), resID, claims.TenantID); err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "checked out successfully"})
}

func (h *Handler) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	propertyID, err := uuid.Parse(r.URL.Query().Get("property_id"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("property_id is required"))
		return
	}

	input := service.AvailabilityInput{
		PropertyID:   propertyID,
		CheckInDate:  r.URL.Query().Get("check_in"),
		CheckOutDate: r.URL.Query().Get("check_out"),
	}

	if input.CheckInDate == "" || input.CheckOutDate == "" {
		response.Err(w, apperrors.BadRequest("check_in and check_out query parameters are required"))
		return
	}

	units, err := h.service.CheckAvailability(r.Context(), claims.TenantID, input)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, units)
}
