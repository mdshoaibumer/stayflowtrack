package domain_test

import (
	"testing"

	"github.com/stayflow/stayflow-track/internal/modules/reservation/domain"
)

func TestReservationStatusTransitions(t *testing.T) {
	tests := []struct {
		name     string
		from     domain.ReservationStatus
		to       domain.ReservationStatus
		expected bool
	}{
		{"pending to confirmed", domain.StatusPending, domain.StatusConfirmed, true},
		{"pending to cancelled", domain.StatusPending, domain.StatusCancelled, true},
		{"pending to checked_in", domain.StatusPending, domain.StatusCheckedIn, false},
		{"confirmed to checked_in", domain.StatusConfirmed, domain.StatusCheckedIn, true},
		{"confirmed to cancelled", domain.StatusConfirmed, domain.StatusCancelled, true},
		{"confirmed to checked_out", domain.StatusConfirmed, domain.StatusCheckedOut, false},
		{"checked_in to checked_out", domain.StatusCheckedIn, domain.StatusCheckedOut, true},
		{"checked_in to cancelled", domain.StatusCheckedIn, domain.StatusCancelled, false},
		{"checked_out to anything", domain.StatusCheckedOut, domain.StatusPending, false},
		{"cancelled to anything", domain.StatusCancelled, domain.StatusPending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.from.CanTransitionTo(tt.to)
			if result != tt.expected {
				t.Errorf("expected %v, got %v for transition %s -> %s",
					tt.expected, result, tt.from, tt.to)
			}
		})
	}
}

func TestReservationStatusIsValid(t *testing.T) {
	tests := []struct {
		status   domain.ReservationStatus
		expected bool
	}{
		{domain.StatusPending, true},
		{domain.StatusConfirmed, true},
		{domain.StatusCheckedIn, true},
		{domain.StatusCheckedOut, true},
		{domain.StatusCancelled, true},
		{domain.ReservationStatus("invalid"), false},
		{domain.ReservationStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if tt.status.IsValid() != tt.expected {
				t.Errorf("expected %v for status %q", tt.expected, tt.status)
			}
		})
	}
}

func TestBookingSourceIsValid(t *testing.T) {
	tests := []struct {
		source   domain.BookingSource
		expected bool
	}{
		{domain.SourceWalkIn, true},
		{domain.SourcePhone, true},
		{domain.SourceWhatsApp, true},
		{domain.SourceBookingCom, true},
		{domain.SourceAirbnb, true},
		{domain.SourceOther, true},
		{domain.BookingSource("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.source), func(t *testing.T) {
			if tt.source.IsValid() != tt.expected {
				t.Errorf("expected %v for source %q", tt.expected, tt.source)
			}
		})
	}
}
