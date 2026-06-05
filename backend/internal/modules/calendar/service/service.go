package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/stayflow/stayflow-track/internal/modules/calendar/domain"
	"github.com/stayflow/stayflow-track/internal/modules/calendar/repository"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetCalendarView(ctx context.Context, tenantID, propertyID uuid.UUID, startDate, endDate string) (*domain.CalendarView, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, apperrors.BadRequest("invalid start_date format, use YYYY-MM-DD")
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, apperrors.BadRequest("invalid end_date format, use YYYY-MM-DD")
	}

	if !end.After(start) {
		return nil, apperrors.BadRequest("end_date must be after start_date")
	}

	// Limit to max 90 days
	if end.Sub(start) > 90*24*time.Hour {
		return nil, apperrors.BadRequest("date range cannot exceed 90 days")
	}

	return s.repo.GetCalendarView(ctx, tenantID, propertyID, start, end)
}

func (s *Service) MoveBooking(ctx context.Context, tenantID uuid.UUID, input domain.MoveBookingInput) error {
	newCheckIn, err := time.Parse("2006-01-02", input.NewCheckIn)
	if err != nil {
		return apperrors.BadRequest("invalid new_check_in format")
	}

	newCheckOut, err := time.Parse("2006-01-02", input.NewCheckOut)
	if err != nil {
		return apperrors.BadRequest("invalid new_check_out format")
	}

	if !newCheckOut.After(newCheckIn) {
		return apperrors.BadRequest("new_check_out must be after new_check_in")
	}

	return s.repo.MoveReservation(ctx, tenantID, input.ReservationID, input.NewUnitID, newCheckIn, newCheckOut)
}

func (s *Service) GetOccupancyStats(ctx context.Context, tenantID, propertyID uuid.UUID, startDate, endDate string) ([]domain.OccupancyStats, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, apperrors.BadRequest("invalid start_date format")
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, apperrors.BadRequest("invalid end_date format")
	}

	if !end.After(start) {
		return nil, apperrors.BadRequest("end_date must be after start_date")
	}

	return s.repo.GetOccupancyStats(ctx, tenantID, propertyID, start, end)
}
