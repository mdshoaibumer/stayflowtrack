package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/stayflow/stayflow-track/internal/modules/property/domain"
	"github.com/stayflow/stayflow-track/internal/modules/property/repository"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

type CreatePropertyInput struct {
	Name    string `json:"name" validate:"required,min=2,max=255"`
	Address string `json:"address" validate:"required"`
	City    string `json:"city" validate:"required,max=100"`
	State   string `json:"state" validate:"required,max=100"`
	Country string `json:"country" validate:"omitempty,max=100"`
	Pincode string `json:"pincode" validate:"omitempty,max=10"`
	Phone   string `json:"phone" validate:"omitempty,max=20"`
	Email   string `json:"email" validate:"omitempty,email"`
}

type UpdatePropertyInput struct {
	Name    string `json:"name" validate:"omitempty,min=2,max=255"`
	Address string `json:"address" validate:"omitempty"`
	City    string `json:"city" validate:"omitempty,max=100"`
	State   string `json:"state" validate:"omitempty,max=100"`
	Country string `json:"country" validate:"omitempty,max=100"`
	Pincode string `json:"pincode" validate:"omitempty,max=10"`
	Phone   string `json:"phone" validate:"omitempty,max=20"`
	Email   string `json:"email" validate:"omitempty,email"`
}

type CreateUnitTypeInput struct {
	Name         string          `json:"name" validate:"required,min=1,max=100"`
	Description  string          `json:"description" validate:"omitempty,max=500"`
	BaseRate     decimal.Decimal `json:"base_rate" validate:"required"`
	MaxOccupancy int             `json:"max_occupancy" validate:"required,min=1,max=20"`
	Amenities    []string        `json:"amenities"`
}

type CreateUnitInput struct {
	UnitTypeID uuid.UUID `json:"unit_type_id" validate:"required"`
	UnitNumber string    `json:"unit_number" validate:"required,min=1,max=20"`
	Floor      string    `json:"floor" validate:"omitempty,max=10"`
	Notes      string    `json:"notes" validate:"omitempty,max=500"`
}

type UpdateUnitInput struct {
	UnitNumber string `json:"unit_number" validate:"omitempty,min=1,max=20"`
	Floor      string `json:"floor" validate:"omitempty,max=10"`
	Notes      string `json:"notes" validate:"omitempty,max=500"`
}

type ChangeStatusInput struct {
	Status string `json:"status" validate:"required,oneof=available reserved occupied cleaning maintenance"`
}

func (s *Service) CreateProperty(ctx context.Context, tenantID uuid.UUID, input CreatePropertyInput) (*domain.Property, error) {
	country := input.Country
	if country == "" {
		country = "India"
	}

	prop := &domain.Property{
		TenantID: tenantID,
		Name:     input.Name,
		Address:  input.Address,
		City:     input.City,
		State:    input.State,
		Country:  country,
		Pincode:  input.Pincode,
		Phone:    input.Phone,
		Email:    input.Email,
	}

	if err := s.repo.CreateProperty(ctx, prop); err != nil {
		return nil, apperrors.Internal(err)
	}

	return prop, nil
}

func (s *Service) GetProperty(ctx context.Context, id, tenantID uuid.UUID) (*domain.Property, error) {
	return s.repo.GetPropertyByID(ctx, id, tenantID)
}

func (s *Service) ListProperties(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Property, int64, error) {
	return s.repo.ListProperties(ctx, tenantID, limit, offset)
}

func (s *Service) UpdateProperty(ctx context.Context, id, tenantID uuid.UUID, input UpdatePropertyInput) (*domain.Property, error) {
	prop := &domain.Property{
		ID:       id,
		TenantID: tenantID,
		Name:     input.Name,
		Address:  input.Address,
		City:     input.City,
		State:    input.State,
		Country:  input.Country,
		Pincode:  input.Pincode,
		Phone:    input.Phone,
		Email:    input.Email,
	}

	if err := s.repo.UpdateProperty(ctx, prop); err != nil {
		return nil, err
	}

	return prop, nil
}

func (s *Service) CreateUnitType(ctx context.Context, tenantID, propertyID uuid.UUID, input CreateUnitTypeInput) (*domain.UnitType, error) {
	// Verify property exists and belongs to tenant
	_, err := s.repo.GetPropertyByID(ctx, propertyID, tenantID)
	if err != nil {
		return nil, err
	}

	ut := &domain.UnitType{
		TenantID:     tenantID,
		PropertyID:   propertyID,
		Name:         input.Name,
		Description:  input.Description,
		BaseRate:     input.BaseRate,
		MaxOccupancy: input.MaxOccupancy,
		Amenities:    input.Amenities,
	}

	if err := s.repo.CreateUnitType(ctx, ut); err != nil {
		return nil, apperrors.Internal(err)
	}

	return ut, nil
}

func (s *Service) ListUnitTypes(ctx context.Context, propertyID, tenantID uuid.UUID) ([]domain.UnitType, error) {
	return s.repo.ListUnitTypesByProperty(ctx, propertyID, tenantID)
}

func (s *Service) CreateUnit(ctx context.Context, tenantID, propertyID uuid.UUID, input CreateUnitInput) (*domain.Unit, error) {
	// Verify property exists
	_, err := s.repo.GetPropertyByID(ctx, propertyID, tenantID)
	if err != nil {
		return nil, err
	}

	unit := &domain.Unit{
		TenantID:   tenantID,
		PropertyID: propertyID,
		UnitTypeID: input.UnitTypeID,
		UnitNumber: input.UnitNumber,
		Floor:      input.Floor,
		Status:     domain.UnitStatusAvailable,
		Notes:      input.Notes,
	}

	if err := s.repo.CreateUnit(ctx, unit); err != nil {
		return nil, apperrors.Internal(err)
	}

	return unit, nil
}

func (s *Service) ListUnits(ctx context.Context, propertyID, tenantID uuid.UUID, limit, offset int) ([]domain.Unit, int64, error) {
	return s.repo.ListUnits(ctx, propertyID, tenantID, limit, offset)
}

func (s *Service) UpdateUnit(ctx context.Context, id, tenantID uuid.UUID, input UpdateUnitInput) (*domain.Unit, error) {
	unit := &domain.Unit{
		ID:         id,
		TenantID:   tenantID,
		UnitNumber: input.UnitNumber,
		Floor:      input.Floor,
		Notes:      input.Notes,
	}

	if err := s.repo.UpdateUnit(ctx, unit); err != nil {
		return nil, err
	}

	return unit, nil
}

func (s *Service) DeleteUnit(ctx context.Context, id, tenantID uuid.UUID) error {
	return s.repo.DeleteUnit(ctx, id, tenantID)
}

func (s *Service) ChangeUnitStatus(ctx context.Context, id, tenantID uuid.UUID, input ChangeStatusInput) error {
	status := domain.UnitStatus(input.Status)
	if !status.IsValid() {
		return apperrors.BadRequest("invalid unit status")
	}

	return s.repo.ChangeUnitStatus(ctx, id, tenantID, status)
}

func (s *Service) SearchUnits(ctx context.Context, params repository.SearchUnitsParams) ([]domain.Unit, int64, error) {
	return s.repo.SearchUnits(ctx, params)
}
