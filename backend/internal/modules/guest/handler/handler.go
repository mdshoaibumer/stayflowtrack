package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/stayflow/stayflow-track/internal/modules/auth/middleware"
	"github.com/stayflow/stayflow-track/internal/modules/guest/service"
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

func (h *Handler) CreateGuest(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var input service.CreateGuestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	guest, err := h.service.CreateGuest(r.Context(), claims.TenantID, input)
	if err != nil {
		h.log.Error().Err(err).Msg("create guest failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, guest)
}

func (h *Handler) GetGuest(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	guestID, err := uuid.Parse(chi.URLParam(r, "guestID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid guest id"))
		return
	}

	guest, err := h.service.GetGuest(r.Context(), guestID, claims.TenantID)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, guest)
}

func (h *Handler) UpdateGuest(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	guestID, err := uuid.Parse(chi.URLParam(r, "guestID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid guest id"))
		return
	}

	var input service.UpdateGuestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, apperrors.BadRequest("invalid request body"))
		return
	}

	if errs := validation.Validate(input); errs != nil {
		response.Err(w, apperrors.Validation("validation failed", errs))
		return
	}

	guest, err := h.service.UpdateGuest(r.Context(), guestID, claims.TenantID, input)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, guest)
}

func (h *Handler) ListGuests(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	pg := pagination.Parse(r)

	guests, total, err := h.service.ListGuests(r.Context(), claims.TenantID, pg.PerPage, pg.Offset)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSONWithMeta(w, http.StatusOK, guests, &response.Meta{
		Page:       pg.Page,
		PerPage:    pg.PerPage,
		Total:      total,
		TotalPages: pagination.TotalPages(total, pg.PerPage),
	})
}

func (h *Handler) SearchGuests(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	pg := pagination.Parse(r)
	query := r.URL.Query().Get("q")

	guests, total, err := h.service.SearchGuests(r.Context(), claims.TenantID, query, pg.PerPage, pg.Offset)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSONWithMeta(w, http.StatusOK, guests, &response.Meta{
		Page:       pg.Page,
		PerPage:    pg.PerPage,
		Total:      total,
		TotalPages: pagination.TotalPages(total, pg.PerPage),
	})
}

func (h *Handler) GetGuestHistory(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	guestID, err := uuid.Parse(chi.URLParam(r, "guestID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid guest id"))
		return
	}

	pg := pagination.Parse(r)

	records, total, err := h.service.GetGuestHistory(r.Context(), guestID, claims.TenantID, pg.PerPage, pg.Offset)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSONWithMeta(w, http.StatusOK, records, &response.Meta{
		Page:       pg.Page,
		PerPage:    pg.PerPage,
		Total:      total,
		TotalPages: pagination.TotalPages(total, pg.PerPage),
	})
}

func (h *Handler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	guestID, err := uuid.Parse(chi.URLParam(r, "guestID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid guest id"))
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		response.Err(w, apperrors.BadRequest("file too large or invalid multipart form"))
		return
	}

	file, header, err := r.FormFile("document")
	if err != nil {
		response.Err(w, apperrors.BadRequest("document file is required"))
		return
	}
	defer file.Close()

	documentType := r.FormValue("document_type")
	if documentType == "" {
		response.Err(w, apperrors.BadRequest("document_type is required"))
		return
	}

	input := service.UploadDocumentInput{
		GuestID:      guestID,
		TenantID:     claims.TenantID,
		DocumentType: documentType,
		FileName:     header.Filename,
		FileSize:     header.Size,
		ContentType:  header.Header.Get("Content-Type"),
		Reader:       file,
	}

	doc, err := h.service.UploadDocument(r.Context(), input)
	if err != nil {
		h.log.Error().Err(err).Msg("upload document failed")
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, doc)
}

func (h *Handler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	guestID, err := uuid.Parse(chi.URLParam(r, "guestID"))
	if err != nil {
		response.Err(w, apperrors.BadRequest("invalid guest id"))
		return
	}

	docs, err := h.service.ListDocuments(r.Context(), guestID, claims.TenantID)
	if err != nil {
		response.Err(w, err)
		return
	}

	response.JSON(w, http.StatusOK, docs)
}
