package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/stayflow/stayflow-track/internal/modules/auth/middleware"
	"github.com/stayflow/stayflow-track/internal/modules/property/repository"
	"github.com/stayflow/stayflow-track/internal/modules/property/service"
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

func (h *Handler) CreateProperty(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input service.CreatePropertyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	prop, err := h.service.CreateProperty(r.Context(), claims.TenantID, input)
	if err != nil {
		h.log.Error().Err(err).Msg("create property failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, prop)
}

func (h *Handler) GetProperty(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	propertyID, err := uuid.Parse(chi.URLParam(r, "propertyID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid property id"))
		return
	}

	prop, err := h.service.GetProperty(r.Context(), propertyID, claims.TenantID)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, prop)
}

func (h *Handler) ListProperties(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	pg := pagination.Parse(r)

	properties, total, err := h.service.ListProperties(r.Context(), claims.TenantID, pg.PerPage, pg.Offset)
	if err != nil {
		h.log.Error().Err(err).Msg("list properties failed")
		response.Err(w, err)
		return
	}

	response.JSONWithMeta(w, http.StatusOK, properties, &response.Meta{
		Page:       pg.Page,
		PerPage:    pg.PerPage,
		Total:      total,
		TotalPages: pagination.TotalPages(total, pg.PerPage),
	})
}

func (h *Handler) UpdateProperty(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	propertyID, err := uuid.Parse(chi.URLParam(r, "propertyID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid property id"))
		return
	}

	var input service.UpdatePropertyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	prop, err := h.service.UpdateProperty(r.Context(), propertyID, claims.TenantID, input)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, prop)
}

func (h *Handler) CreateUnitType(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	propertyID, err := uuid.Parse(chi.URLParam(r, "propertyID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid property id"))
		return
	}

	var input service.CreateUnitTypeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	ut, err := h.service.CreateUnitType(r.Context(), claims.TenantID, propertyID, input)
	if err != nil {
		h.log.Error().Err(err).Msg("create unit type failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, ut)
}

func (h *Handler) ListUnitTypes(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	propertyID, err := uuid.Parse(chi.URLParam(r, "propertyID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid property id"))
		return
	}

	unitTypes, err := h.service.ListUnitTypes(r.Context(), propertyID, claims.TenantID)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, unitTypes)
}

func (h *Handler) CreateUnit(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	propertyID, err := uuid.Parse(chi.URLParam(r, "propertyID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid property id"))
		return
	}

	var input service.CreateUnitInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	unit, err := h.service.CreateUnit(r.Context(), claims.TenantID, propertyID, input)
	if err != nil {
		h.log.Error().Err(err).Msg("create unit failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, unit)
}

func (h *Handler) ListUnits(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	propertyID, err := uuid.Parse(chi.URLParam(r, "propertyID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid property id"))
		return
	}

	pg := pagination.Parse(r)

	units, total, err := h.service.ListUnits(r.Context(), propertyID, claims.TenantID, pg.PerPage, pg.Offset)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSONWithMeta(w, http.StatusOK, units, &response.Meta{
		Page:       pg.Page,
		PerPage:    pg.PerPage,
		Total:      total,
		TotalPages: pagination.TotalPages(total, pg.PerPage),
	})
}

func (h *Handler) UpdateUnit(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	unitID, err := uuid.Parse(chi.URLParam(r, "unitID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid unit id"))
		return
	}

	var input service.UpdateUnitInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	unit, err := h.service.UpdateUnit(r.Context(), unitID, claims.TenantID, input)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, unit)
}

func (h *Handler) DeleteUnit(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	unitID, err := uuid.Parse(chi.URLParam(r, "unitID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid unit id"))
		return
	}

	if err := h.service.DeleteUnit(r.Context(), unitID, claims.TenantID); err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "unit deleted"})
}

func (h *Handler) ChangeUnitStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	unitID, err := uuid.Parse(chi.URLParam(r, "unitID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid unit id"))
		return
	}

	var input service.ChangeStatusInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	if err := h.service.ChangeUnitStatus(r.Context(), unitID, claims.TenantID, input); err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "status updated"})
}

func (h *Handler) SearchUnits(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	propertyID, err := uuid.Parse(chi.URLParam(r, "propertyID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid property id"))
		return
	}

	pg := pagination.Parse(r)

	params := repository.SearchUnitsParams{
		PropertyID: propertyID,
		TenantID:   claims.TenantID,
		Status:     r.URL.Query().Get("status"),
		Floor:      r.URL.Query().Get("floor"),
		Limit:      pg.PerPage,
		Offset:     pg.Offset,
	}

	if utID := r.URL.Query().Get("unit_type_id"); utID != "" {
		id, err := uuid.Parse(utID)
		if err == nil {
			params.UnitTypeID = &id
		}
	}

	units, total, err := h.service.SearchUnits(r.Context(), params)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSONWithMeta(w, http.StatusOK, units, &response.Meta{
		Page:       pg.Page,
		PerPage:    pg.PerPage,
		Total:      total,
		TotalPages: pagination.TotalPages(total, pg.PerPage),
	})
}
