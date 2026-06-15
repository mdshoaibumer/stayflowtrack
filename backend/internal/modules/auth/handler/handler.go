package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/stayflow/stayflow-track/internal/modules/auth/middleware"
	"github.com/stayflow/stayflow-track/internal/modules/auth/service"
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

func (h *Handler) RegisterTenant(w http.ResponseWriter, r *http.Request) {
	var input service.RegisterTenantInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	result, err := h.service.RegisterTenant(r.Context(), input)
	if err != nil {
		h.log.Error().Err(err).Msg("register tenant failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, result)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input service.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	tokens, err := h.service.Login(r.Context(), input)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, tokens)
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var input service.RefreshInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	tokens, err := h.service.RefreshToken(r.Context(), input)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, tokens)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		response.Err(w, apperrors.Unauthorized("missing authentication"))
		return
	}

	if err := h.service.Logout(r.Context(), claims.UserID); err != nil {
		h.log.Error().Err(err).Msg("logout failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		response.Err(w, apperrors.Unauthorized("missing authentication"))
		return
	}

	var input service.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	user, err := h.service.CreateUser(r.Context(), claims.TenantID, claims.RoleName, input)
	if err != nil {
		h.log.Error().Err(err).Msg("create user failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, user)
}

func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var input service.PasswordResetRequestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	_, err := h.service.RequestPasswordReset(r.Context(), input)
	if err != nil {
		h.log.Error().Err(err).Msg("password reset request failed")
	}

	// Always return success to prevent email enumeration
	response.JSON(w, http.StatusOK, map[string]string{
		"message": "if an account with that email exists, a password reset link has been sent",
	})
}

func (h *Handler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var input service.PasswordResetConfirmInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	if err := h.service.ConfirmPasswordReset(r.Context(), input); err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "password reset successful"})
}
