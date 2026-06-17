package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	notifservice "github.com/stayflow/stayflow-track/internal/modules/notifications/service"
	"github.com/stayflow/stayflow-track/internal/modules/reservation/domain"
	resrepo "github.com/stayflow/stayflow-track/internal/modules/reservation/repository"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

type Service struct {
	repo     *resrepo.Repository
	propRepo PropertyReader
	notifSvc *notifservice.Service
}

func New(repo *resrepo.Repository, propRepo PropertyReader, notifSvc *notifservice.Service) *Service {
	return &Service{repo: repo, propRepo: propRepo, notifSvc: notifSvc}
}

type CreateReservationInput struct {
	PropertyID        uuid.UUID       `json:"property_id" validate:"required"`
	UnitID            uuid.UUID       `json:"unit_id" validate:"required"`
	GuestID           uuid.UUID       `json:"guest_id" validate:"required"`
	BookingSource     string          `json:"booking_source" validate:"required,oneof=walk_in phone whatsapp email website booking_com airbnb ota_makemytrip ota_goibibo ota_other referral corporate repeat other"`
	CheckInDate       string          `json:"check_in_date" validate:"required"`
	CheckOutDate      string          `json:"check_out_date" validate:"required"`
	NumGuests         int             `json:"num_guests" validate:"required,min=1,max=20"`
	RatePerNight      decimal.Decimal `json:"rate_per_night" validate:"required"`
	Notes             string          `json:"notes" validate:"omitempty,max=1000"`
	ExternalBookingID string          `json:"external_booking_id" validate:"omitempty,max=255"`
	AdvanceAmount     decimal.Decimal `json:"advance_amount"`
	AdvanceMethod     string          `json:"advance_method" validate:"omitempty,oneof=cash upi card bank_transfer cheque"`
	AdvanceReference  string          `json:"advance_reference" validate:"omitempty,max=255"`
}

type UpdateReservationInput struct {
	CheckInDate  string          `json:"check_in_date" validate:"omitempty"`
	CheckOutDate string          `json:"check_out_date" validate:"omitempty"`
	NumGuests    int             `json:"num_guests" validate:"omitempty,min=1,max=20"`
	RatePerNight decimal.Decimal `json:"rate_per_night" validate:"omitempty"`
	Notes        string          `json:"notes" validate:"omitempty,max=1000"`
}

type CancelReservationInput struct {
	Reason string `json:"reason" validate:"required,min=3,max=500"`
}

type AvailabilityInput struct {
	PropertyID   uuid.UUID `json:"property_id" validate:"required"`
	CheckInDate  string    `json:"check_in_date" validate:"required"`
	CheckOutDate string    `json:"check_out_date" validate:"required"`
}

func (s *Service) CreateReservation(ctx context.Context, tenantID uuid.UUID, input CreateReservationInput) (*domain.Reservation, error) {
	checkIn, err := time.Parse("2006-01-02", input.CheckInDate)
	if err != nil {
		return nil, apperrors.BadRequest("invalid check_in_date format, use YYYY-MM-DD")
	}

	checkOut, err := time.Parse("2006-01-02", input.CheckOutDate)
	if err != nil {
		return nil, apperrors.BadRequest("invalid check_out_date format, use YYYY-MM-DD")
	}

	if !checkOut.After(checkIn) {
		return nil, apperrors.BadRequest("check_out_date must be after check_in_date")
	}

	if checkIn.Before(time.Now().Truncate(24 * time.Hour)) {
		return nil, apperrors.BadRequest("check_in_date cannot be in the past")
	}

	// Verify unit exists and belongs to property
	unit, err := s.propRepo.GetUnitByID(ctx, input.UnitID, tenantID)
	if err != nil {
		return nil, err
	}
	if unit.PropertyID != input.PropertyID {
		return nil, apperrors.BadRequest("unit does not belong to the specified property")
	}

	// Calculate total amount using decimal precision
	nights := int(math.Ceil(checkOut.Sub(checkIn).Hours() / 24))
	totalAmount := decimal.NewFromInt(int64(nights)).Mul(input.RatePerNight)

	res := &domain.Reservation{
		TenantID:          tenantID,
		PropertyID:        input.PropertyID,
		UnitID:            input.UnitID,
		GuestID:           input.GuestID,
		BookingSource:     domain.BookingSource(input.BookingSource),
		Status:            domain.StatusPending,
		CheckInDate:       checkIn,
		CheckOutDate:      checkOut,
		NumGuests:         input.NumGuests,
		RatePerNight:      input.RatePerNight,
		TotalAmount:       totalAmount,
		AdvanceAmount:     input.AdvanceAmount,
		AdvanceMethod:     input.AdvanceMethod,
		AdvanceReference:  input.AdvanceReference,
		Notes:             input.Notes,
		ExternalBookingID: input.ExternalBookingID,
	}

	// Atomic conflict check + insert within a single transaction to prevent double-booking
	if err := s.repo.CreateReservationAtomic(ctx, res); err != nil {
		return nil, err
	}

	// Update unit status to reserved (with tenant isolation, log errors)
	if err := s.repo.UpdateUnitStatusWithTenant(ctx, input.UnitID, tenantID, "reserved"); err != nil {
		// Log but don't fail the reservation — unit status is secondary
		log.Warn().Err(err).Str("unit_id", input.UnitID.String()).Msg("failed to update unit status to reserved")
	}

	// Send booking confirmation notification (async, non-blocking)
	if s.notifSvc != nil {
		resID := res.ID
		go func() {
			defer func() {
				if r := recover(); r != nil {
					// Log panic but don't crash the process
				}
			}()
			// Look up full reservation data with joined guest/property fields
			full, err := s.repo.GetReservationByID(context.Background(), resID, tenantID)
			if err != nil {
				return
			}
			guestName := (full.GuestFirstName + " " + full.GuestLastName)
			s.notifSvc.SendBookingConfirmation(context.Background(), tenantID,
				full.GuestPhone,
				guestName,
				full.PropertyName,
				input.CheckInDate,
				input.CheckOutDate,
				resID.String(),
			)
		}()
	}

	return res, nil
}

func (s *Service) GetReservation(ctx context.Context, id, tenantID uuid.UUID) (*domain.Reservation, error) {
	return s.repo.GetReservationByID(ctx, id, tenantID)
}

func (s *Service) ListReservations(ctx context.Context, params resrepo.ListParams) ([]domain.Reservation, int64, error) {
	return s.repo.ListReservations(ctx, params)
}

func (s *Service) UpdateReservation(ctx context.Context, id, tenantID uuid.UUID, input UpdateReservationInput) (*domain.Reservation, error) {
	existing, err := s.repo.GetReservationByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	if existing.Status != domain.StatusPending && existing.Status != domain.StatusConfirmed {
		return nil, apperrors.BadRequest("only pending or confirmed reservations can be updated")
	}

	checkIn := existing.CheckInDate
	checkOut := existing.CheckOutDate

	if input.CheckInDate != "" {
		checkIn, err = time.Parse("2006-01-02", input.CheckInDate)
		if err != nil {
			return nil, apperrors.BadRequest("invalid check_in_date format")
		}
	}

	if input.CheckOutDate != "" {
		checkOut, err = time.Parse("2006-01-02", input.CheckOutDate)
		if err != nil {
			return nil, apperrors.BadRequest("invalid check_out_date format")
		}
	}

	if !checkOut.After(checkIn) {
		return nil, apperrors.BadRequest("check_out_date must be after check_in_date")
	}

	// Check conflicts if dates changed (atomic with update to prevent TOCTOU race)
	if checkIn != existing.CheckInDate || checkOut != existing.CheckOutDate {
		hasConflict, err := s.repo.CheckConflictForUpdate(ctx, existing.UnitID, checkIn, checkOut, &id, tenantID)
		if err != nil {
			return nil, apperrors.Internal(err)
		}
		if hasConflict {
			return nil, apperrors.Conflict("unit is not available for the updated dates")
		}
	}

	ratePerNight := existing.RatePerNight
	if input.RatePerNight.IsPositive() {
		ratePerNight = input.RatePerNight
	}

	numGuests := existing.NumGuests
	if input.NumGuests > 0 {
		numGuests = input.NumGuests
	}

	notes := existing.Notes
	if input.Notes != "" {
		notes = input.Notes
	}

	nights := int(math.Ceil(checkOut.Sub(checkIn).Hours() / 24))
	totalAmount := decimal.NewFromInt(int64(nights)).Mul(ratePerNight)

	res := &domain.Reservation{
		ID:           id,
		TenantID:     tenantID,
		CheckInDate:  checkIn,
		CheckOutDate: checkOut,
		NumGuests:    numGuests,
		RatePerNight: ratePerNight,
		TotalAmount:  totalAmount,
		Notes:        notes,
	}

	if err := s.repo.UpdateReservation(ctx, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Service) CancelReservation(ctx context.Context, id, tenantID uuid.UUID, input CancelReservationInput) error {
	existing, err := s.repo.GetReservationByID(ctx, id, tenantID)
	if err != nil {
		return err
	}

	if !existing.Status.CanTransitionTo(domain.StatusCancelled) {
		return apperrors.BadRequest(fmt.Sprintf("cannot cancel reservation in '%s' status", existing.Status))
	}

	if err := s.repo.CancelReservation(ctx, id, tenantID, input.Reason); err != nil {
		return err
	}

	// Release the unit (with tenant isolation)
	if err := s.repo.UpdateUnitStatusWithTenant(ctx, existing.UnitID, tenantID, "available"); err != nil {
		log.Warn().Err(err).Str("unit_id", existing.UnitID.String()).Str("reservation_id", id.String()).Msg("failed to update unit status to available on cancel")
	}

	return nil
}

// ConfirmReservation transitions a pending reservation to confirmed.
func (s *Service) ConfirmReservation(ctx context.Context, id, tenantID uuid.UUID) error {
	existing, err := s.repo.GetReservationByID(ctx, id, tenantID)
	if err != nil {
		return err
	}

	if !existing.Status.CanTransitionTo(domain.StatusConfirmed) {
		return apperrors.BadRequest(fmt.Sprintf("cannot confirm reservation in '%s' status", existing.Status))
	}

	return s.repo.ConfirmReservation(ctx, id, tenantID)
}

func (s *Service) CheckIn(ctx context.Context, id, tenantID uuid.UUID) error {
	existing, err := s.repo.GetReservationByID(ctx, id, tenantID)
	if err != nil {
		return err
	}

	if !existing.Status.CanTransitionTo(domain.StatusCheckedIn) {
		return apperrors.BadRequest(fmt.Sprintf("cannot check in reservation in '%s' status", existing.Status))
	}

	if err := s.repo.CheckIn(ctx, id, tenantID); err != nil {
		return err
	}

	// Update unit status to occupied (with tenant isolation)
	if err := s.repo.UpdateUnitStatusWithTenant(ctx, existing.UnitID, tenantID, "occupied"); err != nil {
		log.Warn().Err(err).Str("unit_id", existing.UnitID.String()).Str("reservation_id", id.String()).Msg("failed to update unit status to occupied on check-in")
	}

	return nil
}

func (s *Service) CheckOut(ctx context.Context, id, tenantID uuid.UUID) error {
	existing, err := s.repo.GetReservationByID(ctx, id, tenantID)
	if err != nil {
		return err
	}

	if !existing.Status.CanTransitionTo(domain.StatusCheckedOut) {
		return apperrors.BadRequest(fmt.Sprintf("cannot check out reservation in '%s' status", existing.Status))
	}

	if err := s.repo.CheckOut(ctx, id, tenantID); err != nil {
		return err
	}

	// Update unit status to cleaning after checkout (with tenant isolation)
	if err := s.repo.UpdateUnitStatusWithTenant(ctx, existing.UnitID, tenantID, "cleaning"); err != nil {
		log.Warn().Err(err).Str("unit_id", existing.UnitID.String()).Str("reservation_id", id.String()).Msg("failed to update unit status to cleaning on checkout")
	}

	// Note: guest stay count is incremented in the checkinout module's PerformCheckOut transaction.
	// Do NOT increment here to avoid double-counting.

	return nil
}

func (s *Service) CheckAvailability(ctx context.Context, tenantID uuid.UUID, input AvailabilityInput) ([]domain.AvailableUnit, error) {
	checkIn, err := time.Parse("2006-01-02", input.CheckInDate)
	if err != nil {
		return nil, apperrors.BadRequest("invalid check_in_date format, use YYYY-MM-DD")
	}

	checkOut, err := time.Parse("2006-01-02", input.CheckOutDate)
	if err != nil {
		return nil, apperrors.BadRequest("invalid check_out_date format, use YYYY-MM-DD")
	}

	if !checkOut.After(checkIn) {
		return nil, apperrors.BadRequest("check_out_date must be after check_in_date")
	}

	return s.repo.GetAvailableUnits(ctx, input.PropertyID, tenantID, checkIn, checkOut)
}
