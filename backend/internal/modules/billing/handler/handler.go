package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/stayflow/stayflow-track/internal/modules/auth/middleware"
	"github.com/stayflow/stayflow-track/internal/modules/billing/service"
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

func (h *Handler) GetFolio(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	folioID, err := uuid.Parse(chi.URLParam(r, "folioID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid folio id"))
		return
	}

	folio, err := h.service.GetFolio(r.Context(), folioID, claims.TenantID)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, folio)
}

func (h *Handler) GetFolioByReservation(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	reservationID, err := uuid.Parse(chi.URLParam(r, "reservationID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid reservation id"))
		return
	}

	folio, err := h.service.GetFolioByReservation(r.Context(), reservationID, claims.TenantID)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, folio)
}

func (h *Handler) GetFolioSummary(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	folioID, err := uuid.Parse(chi.URLParam(r, "folioID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid folio id"))
		return
	}

	summary, err := h.service.GetFolioSummary(r.Context(), folioID, claims.TenantID)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, summary)
}

func (h *Handler) AddCharge(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input service.AddChargeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	item, err := h.service.AddCharge(r.Context(), claims.TenantID, claims.UserID, input)
	if err != nil {
		h.log.Error().Err(err).Msg("add charge failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, item)
}

func (h *Handler) VoidLineItem(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input service.VoidItemInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	if err := h.service.VoidLineItem(r.Context(), claims.TenantID, input); err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "line item voided"})
}

func (h *Handler) ListLineItems(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	folioID, err := uuid.Parse(chi.URLParam(r, "folioID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid folio id"))
		return
	}

	items, err := h.service.ListLineItems(r.Context(), folioID, claims.TenantID)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, items)
}

func (h *Handler) RecordPayment(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input service.RecordPaymentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	payment, err := h.service.RecordPayment(r.Context(), claims.TenantID, claims.UserID, input)
	if err != nil {
		h.log.Error().Err(err).Msg("record payment failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, payment)
}

func (h *Handler) ListPayments(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	folioID, err := uuid.Parse(chi.URLParam(r, "folioID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid folio id"))
		return
	}

	payments, err := h.service.ListPayments(r.Context(), folioID, claims.TenantID)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, payments)
}

func (h *Handler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	invoiceID, err := uuid.Parse(chi.URLParam(r, "invoiceID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid invoice id"))
		return
	}

	invoice, err := h.service.GetInvoice(r.Context(), invoiceID, claims.TenantID)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, invoice)
}

func (h *Handler) ListInvoices(w http.ResponseWriter, r *http.Request) {
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

	invoices, total, err := h.service.ListInvoices(r.Context(), claims.TenantID, propertyID, status, pg.PerPage, pg.Offset)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSONWithMeta(w, http.StatusOK, invoices, &response.Meta{
		Page:       pg.Page,
		PerPage:    pg.PerPage,
		Total:      total,
		TotalPages: pagination.TotalPages(total, pg.PerPage),
	})
}

func (h *Handler) GenerateInvoicePDF(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	invoiceID, err := uuid.Parse(chi.URLParam(r, "invoiceID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid invoice id"))
		return
	}

	url, err := h.service.GenerateInvoicePDF(r.Context(), invoiceID, claims.TenantID)
	if err != nil {
		h.log.Error().Err(err).Msg("generate invoice pdf failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"pdf_url": url})
}

func (h *Handler) GetInvoicePDF(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	invoiceID, err := uuid.Parse(chi.URLParam(r, "invoiceID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid invoice id"))
		return
	}

	url, err := h.service.GetInvoicePDFURL(r.Context(), invoiceID, claims.TenantID)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"pdf_url": url})
}
