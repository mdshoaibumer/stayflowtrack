package service

import (
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/google/uuid"

	"github.com/stayflow/stayflow-track/internal/modules/guest/domain"
	"github.com/stayflow/stayflow-track/internal/modules/guest/repository"
	"github.com/stayflow/stayflow-track/internal/platform/storage"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

type Service struct {
	repo  *repository.Repository
	store storage.Store
}

func New(repo *repository.Repository, store storage.Store) *Service {
	return &Service{repo: repo, store: store}
}

type CreateGuestInput struct {
	FirstName      string     `json:"first_name" validate:"required,min=1,max=100"`
	LastName       string     `json:"last_name" validate:"required,min=1,max=100"`
	Email          string     `json:"email" validate:"omitempty,email"`
	Phone          string     `json:"phone" validate:"required,min=10,max=20"`
	Address        string     `json:"address" validate:"omitempty,max=500"`
	City           string     `json:"city" validate:"omitempty,max=100"`
	State          string     `json:"state" validate:"omitempty,max=100"`
	Country        string     `json:"country" validate:"omitempty,max=100"`
	Pincode        string     `json:"pincode" validate:"omitempty,max=10"`
	Nationality    string     `json:"nationality" validate:"omitempty,max=100"`
	DateOfBirth    *time.Time `json:"date_of_birth"`
	AadhaarNumber  string     `json:"aadhaar_number" validate:"omitempty,len=12"`
	PassportNumber string     `json:"passport_number" validate:"omitempty,max=20"`
	Notes          string     `json:"notes" validate:"omitempty,max=1000"`
}

type UpdateGuestInput struct {
	FirstName      string     `json:"first_name" validate:"omitempty,min=1,max=100"`
	LastName       string     `json:"last_name" validate:"omitempty,min=1,max=100"`
	Email          string     `json:"email" validate:"omitempty,email"`
	Phone          string     `json:"phone" validate:"omitempty,min=10,max=20"`
	Address        string     `json:"address" validate:"omitempty,max=500"`
	City           string     `json:"city" validate:"omitempty,max=100"`
	State          string     `json:"state" validate:"omitempty,max=100"`
	Country        string     `json:"country" validate:"omitempty,max=100"`
	Pincode        string     `json:"pincode" validate:"omitempty,max=10"`
	Nationality    string     `json:"nationality" validate:"omitempty,max=100"`
	DateOfBirth    *time.Time `json:"date_of_birth"`
	AadhaarNumber  string     `json:"aadhaar_number" validate:"omitempty,len=12"`
	PassportNumber string     `json:"passport_number" validate:"omitempty,max=20"`
	Notes          string     `json:"notes" validate:"omitempty,max=1000"`
}

type UploadDocumentInput struct {
	GuestID      uuid.UUID
	TenantID     uuid.UUID
	DocumentType string
	FileName     string
	FileSize     int64
	ContentType  string
	Reader       io.Reader
}

var allowedDocumentTypes = map[string]bool{
	"aadhaar":         true,
	"passport":        true,
	"driving_license": true,
	"voter_id":        true,
	"other":           true,
}

var allowedContentTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
	"application/pdf": true,
}

const maxFileSize = 10 * 1024 * 1024 // 10MB

func (s *Service) CreateGuest(ctx context.Context, tenantID uuid.UUID, input CreateGuestInput) (*domain.Guest, error) {
	guest := &domain.Guest{
		TenantID:       tenantID,
		FirstName:      input.FirstName,
		LastName:       input.LastName,
		Email:          input.Email,
		Phone:          input.Phone,
		Address:        input.Address,
		City:           input.City,
		State:          input.State,
		Country:        input.Country,
		Pincode:        input.Pincode,
		Nationality:    input.Nationality,
		DateOfBirth:    input.DateOfBirth,
		AadhaarNumber:  input.AadhaarNumber,
		PassportNumber: input.PassportNumber,
		Notes:          input.Notes,
	}

	if err := s.repo.CreateGuest(ctx, guest); err != nil {
		return nil, apperrors.Internal(err)
	}

	return guest, nil
}

func (s *Service) GetGuest(ctx context.Context, id, tenantID uuid.UUID) (*domain.Guest, error) {
	return s.repo.GetGuestByID(ctx, id, tenantID)
}

func (s *Service) UpdateGuest(ctx context.Context, id, tenantID uuid.UUID, input UpdateGuestInput) (*domain.Guest, error) {
	guest := &domain.Guest{
		ID:             id,
		TenantID:       tenantID,
		FirstName:      input.FirstName,
		LastName:       input.LastName,
		Email:          input.Email,
		Phone:          input.Phone,
		Address:        input.Address,
		City:           input.City,
		State:          input.State,
		Country:        input.Country,
		Pincode:        input.Pincode,
		Nationality:    input.Nationality,
		DateOfBirth:    input.DateOfBirth,
		AadhaarNumber:  input.AadhaarNumber,
		PassportNumber: input.PassportNumber,
		Notes:          input.Notes,
	}

	if err := s.repo.UpdateGuest(ctx, guest); err != nil {
		return nil, err
	}

	return guest, nil
}

func (s *Service) ListGuests(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Guest, int64, error) {
	return s.repo.ListGuests(ctx, tenantID, limit, offset)
}

func (s *Service) SearchGuests(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]domain.Guest, int64, error) {
	return s.repo.SearchGuests(ctx, tenantID, query, limit, offset)
}

func (s *Service) GetGuestHistory(ctx context.Context, guestID, tenantID uuid.UUID, limit, offset int) ([]repository.GuestStayRecord, int64, error) {
	// Verify guest exists
	_, err := s.repo.GetGuestByID(ctx, guestID, tenantID)
	if err != nil {
		return nil, 0, err
	}

	return s.repo.GetGuestHistory(ctx, guestID, tenantID, limit, offset)
}

func (s *Service) UploadDocument(ctx context.Context, input UploadDocumentInput) (*domain.GuestDocument, error) {
	if !allowedDocumentTypes[input.DocumentType] {
		return nil, apperrors.BadRequest("invalid document type")
	}

	if !allowedContentTypes[input.ContentType] {
		return nil, apperrors.BadRequest("unsupported file type, allowed: jpeg, png, webp, pdf")
	}

	if input.FileSize > maxFileSize {
		return nil, apperrors.BadRequest("file size exceeds 10MB limit")
	}

	// Verify guest exists
	_, err := s.repo.GetGuestByID(ctx, input.GuestID, input.TenantID)
	if err != nil {
		return nil, err
	}

	// Generate storage key
	key := fmt.Sprintf("tenants/%s/guests/%s/documents/%s%s",
		input.TenantID.String(),
		input.GuestID.String(),
		uuid.New().String(),
		path.Ext(input.FileName),
	)

	_, err = s.store.Upload(ctx, key, input.Reader, input.ContentType, input.FileSize)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("upload document: %w", err))
	}

	doc := &domain.GuestDocument{
		TenantID:     input.TenantID,
		GuestID:      input.GuestID,
		DocumentType: input.DocumentType,
		FileKey:      key,
		FileName:     input.FileName,
		FileSize:     input.FileSize,
		ContentType:  input.ContentType,
	}

	if err := s.repo.CreateGuestDocument(ctx, doc); err != nil {
		// Attempt cleanup on failure
		_ = s.store.Delete(ctx, key)
		return nil, apperrors.Internal(err)
	}

	return doc, nil
}

func (s *Service) ListDocuments(ctx context.Context, guestID, tenantID uuid.UUID) ([]domain.GuestDocument, error) {
	docs, err := s.repo.ListGuestDocuments(ctx, guestID, tenantID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}

	// Generate presigned URLs
	for i := range docs {
		url, err := s.store.GetPresignedURL(ctx, docs[i].FileKey, 15*time.Minute)
		if err == nil {
			docs[i].DownloadURL = url
		}
	}

	return docs, nil
}
