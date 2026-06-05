package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/stayflow/stayflow-track/internal/modules/auth/middleware"
	"github.com/stayflow/stayflow-track/internal/modules/operations/domain"
	"github.com/stayflow/stayflow-track/internal/modules/operations/service"
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

func (h *Handler) MarkNoShow(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input domain.NoShowInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	if err := h.service.MarkNoShow(r.Context(), claims.TenantID, input); err != nil {
		h.log.Error().Err(err).Msg("mark no-show failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "reservation marked as no-show"})
}

func (h *Handler) ExtendStay(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input domain.ExtendStayInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	ext, err := h.service.ExtendStay(r.Context(), claims.TenantID, claims.UserID, input)
	if err != nil {
		h.log.Error().Err(err).Msg("extend stay failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, ext)
}

func (h *Handler) MoveRoom(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input domain.RoomMoveInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	move, err := h.service.MoveRoom(r.Context(), claims.TenantID, claims.UserID, input)
	if err != nil {
		h.log.Error().Err(err).Msg("room move failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, move)
}

func (h *Handler) CreateMaintenanceBlock(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input domain.MaintenanceBlockInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	block, err := h.service.CreateMaintenanceBlock(r.Context(), claims.TenantID, claims.UserID, input)
	if err != nil {
		h.log.Error().Err(err).Msg("create maintenance block failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, block)
}

func (h *Handler) RefundDeposit(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input domain.RefundDepositInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	if err := h.service.RefundDeposit(r.Context(), claims.TenantID, claims.UserID, input); err != nil {
		h.log.Error().Err(err).Msg("refund deposit failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "deposit refunded successfully"})
}
