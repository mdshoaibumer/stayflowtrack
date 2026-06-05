package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/stayflow/stayflow-track/internal/modules/auth/middleware"
	"github.com/stayflow/stayflow-track/internal/modules/laundry/domain"
	"github.com/stayflow/stayflow-track/internal/modules/laundry/service"
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

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input domain.CreateOrderInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	order, err := h.service.CreateOrder(r.Context(), claims.TenantID, claims.UserID, input)
	if err != nil {
		h.log.Error().Err(err).Msg("create laundry order failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, order)
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	orderID, err := uuid.Parse(chi.URLParam(r, "orderID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid order id"))
		return
	}

	order, err := h.service.GetOrder(r.Context(), orderID, claims.TenantID)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, order)
}

func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	pg := pagination.Parse(r)

	var propertyID *uuid.UUID
	if pid := r.URL.Query().Get("property_id"); pid != "" {
		id, err := uuid.Parse(pid)
		if err == nil {
			propertyID = &id
		}
	}
	status := r.URL.Query().Get("status")

	orders, total, err := h.service.ListOrders(r.Context(), claims.TenantID, propertyID, status, pg.PerPage, pg.Offset)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSONWithMeta(w, http.StatusOK, orders, &response.Meta{
		Page:       pg.Page,
		PerPage:    pg.PerPage,
		Total:      total,
		TotalPages: pagination.TotalPages(total, pg.PerPage),
	})
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input domain.UpdateStatusInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	if err := h.service.UpdateStatus(r.Context(), claims.TenantID, claims.UserID, input); err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "status updated"})
}

func (h *Handler) PostToFolio(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	orderID, err := uuid.Parse(chi.URLParam(r, "orderID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid order id"))
		return
	}

	if err := h.service.PostToFolio(r.Context(), orderID, claims.TenantID, claims.UserID); err != nil {
		h.log.Error().Err(err).Msg("post laundry to folio failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "charges posted to folio"})
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	propertyID, err := uuid.Parse(chi.URLParam(r, "propertyID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid property id"))
		return
	}

	stats, err := h.service.GetStats(r.Context(), claims.TenantID, propertyID)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, stats)
}
