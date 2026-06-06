package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/stayflow/stayflow-track/internal/modules/checkinout/domain"
	"github.com/stayflow/stayflow-track/internal/modules/checkinout/repository"
	notifservice "github.com/stayflow/stayflow-track/internal/modules/notifications/service"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

// NotificationSender is an optional interface for triggering notifications.
type NotificationSender interface {
	SendBookingConfirmation(ctx context.Context, tenantID uuid.UUID, phone, guestName, propertyName, checkIn, checkOut, confirmationNo string)
}

type Service struct {
	repo     *repository.Repository
	notifSvc *notifservice.Service
}

func New(repo *repository.Repository, notifSvc *notifservice.Service) *Service {
	return &Service{repo: repo, notifSvc: notifSvc}
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

	// Deposit is now mandatory - validated by struct tags
	// ID document is now mandatory - validated by struct tags

	// Perform atomic check-in
	details, folioID, err := s.repo.PerformCheckIn(ctx, tenantID, input, info, userID)
	if err != nil {
		return nil, err
	}

	// Send check-in confirmation notification (async, don't block on failure)
	if s.notifSvc != nil {
		go s.notifSvc.SendBookingConfirmation(context.Background(), tenantID,
			"",                 // phone fetched from guest record
			info.GuestName, "", // property name
			info.CheckInDate.Format("02-Jan-2006"),
			info.CheckOutDate.Format("02-Jan-2006"),
			info.ID.String(),
		)
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

	// Auto-detect late checkout and suggest charge if not provided
	if input.LateCheckoutCharge.IsZero() {
		expectedCheckout := time.Date(info.CheckOutDate.Year(), info.CheckOutDate.Month(), info.CheckOutDate.Day(), 11, 0, 0, 0, time.Local)
		if time.Now().After(expectedCheckout) {
			lateHours := time.Since(expectedCheckout).Hours()
			// Auto-calculate: half-day rate for 1-4 hours, full-day for 4+ hours
			if lateHours > 4 {
				input.LateCheckoutCharge = info.RatePerNight
			} else if lateHours > 1 {
				input.LateCheckoutCharge = info.RatePerNight.Div(decimal.NewFromInt(2)).Round(0)
			}
		}
	}

	result, err := s.repo.PerformCheckOut(ctx, tenantID, info, userID, input.Notes, input.LateCheckoutCharge)
	if err != nil {
		return nil, err
	}

	// Auto-create housekeeping task for checkout cleaning
	go s.repo.CreateCheckoutCleaningTask(context.Background(), tenantID, info.UnitID, info.PropertyID, info.UnitNumber)

	return result, nil
}

// WalkIn creates a guest + reservation + check-in in one atomic operation.
func (s *Service) WalkIn(ctx context.Context, tenantID, userID uuid.UUID, input domain.WalkInInput) (*domain.WalkInResult, error) {
	checkOut, err := time.Parse("2006-01-02", input.CheckOutDate)
	if err != nil {
		return nil, apperrors.BadRequest("invalid check_out_date format, use YYYY-MM-DD")
	}

	today := time.Now().Truncate(24 * time.Hour)
	if checkOut.Before(today) || checkOut.Equal(today) {
		return nil, apperrors.BadRequest("check_out_date must be after today")
	}

	result, err := s.repo.PerformWalkIn(ctx, tenantID, userID, input, checkOut)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetPreCheckoutSummary returns the full bill before checkout confirmation.
func (s *Service) GetPreCheckoutSummary(ctx context.Context, tenantID, reservationID uuid.UUID) (*domain.FolioSummary, error) {
	info, err := s.repo.GetReservationForCheckIn(ctx, reservationID, tenantID)
	if err != nil {
		return nil, err
	}

	if info.Status != "checked_in" {
		return nil, apperrors.BadRequest("reservation must be in 'checked_in' status")
	}

	summary, err := s.repo.GetFolioSummary(ctx, tenantID, info)
	if err != nil {
		return nil, err
	}

	return summary, nil
}
