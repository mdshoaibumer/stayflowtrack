package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/stayflow/stayflow-track/internal/modules/checkinout/domain"
	"github.com/stayflow/stayflow-track/internal/modules/checkinout/repository"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

type CheckInResult struct {
	Details *domain.CheckInDetails `json:"details"`
	FolioID uuid.UUID              `json:"folio_id"`
}

func (s *Service) CheckIn(ctx context.Context, tenantID, userID uuid.UUID, input domain.CheckInInput) (*CheckInResult, error) {
	// Validate reservation exists and is in correct state
	info, err := s.repo.GetReservationForCheckIn(ctx, input.ReservationID, tenantID)
	if err != nil {
		return nil, err
	}

	if info.Status != "confirmed" {
		return nil, apperrors.BadRequest("reservation must be in 'confirmed' status to check in")
	}

	// If deposit amount is provided, method must also be provided
	if input.DepositAmount.IsPositive() && input.DepositMethod == "" {
		return nil, apperrors.BadRequest("deposit_method is required when deposit_amount is provided")
	}

	// Perform atomic check-in
	details, folioID, err := s.repo.PerformCheckIn(ctx, tenantID, input, info, userID)
	if err != nil {
		return nil, err
	}

	return &CheckInResult{
		Details: details,
		FolioID: folioID,
	}, nil
}

func (s *Service) CheckOut(ctx context.Context, tenantID, userID uuid.UUID, input domain.CheckOutInput) (*domain.CheckOutResult, error) {
	info, err := s.repo.GetReservationForCheckIn(ctx, input.ReservationID, tenantID)
	if err != nil {
		return nil, err
	}

	if info.Status != "checked_in" {
		return nil, apperrors.BadRequest("reservation must be in 'checked_in' status to check out")
	}

	result, err := s.repo.PerformCheckOut(ctx, tenantID, info, userID, input.Notes)
	if err != nil {
		return nil, err
	}

	return result, nil
}
