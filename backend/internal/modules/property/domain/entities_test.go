package domain_test

import (
	"testing"

	"github.com/stayflow/stayflow-track/internal/modules/property/domain"
)

func TestUnitStatusIsValid(t *testing.T) {
	tests := []struct {
		status   domain.UnitStatus
		expected bool
	}{
		{domain.UnitStatusAvailable, true},
		{domain.UnitStatusReserved, true},
		{domain.UnitStatusOccupied, true},
		{domain.UnitStatusCleaning, true},
		{domain.UnitStatusMaintenance, true},
		{domain.UnitStatus("invalid"), false},
		{domain.UnitStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if tt.status.IsValid() != tt.expected {
				t.Errorf("expected %v for status %q", tt.expected, tt.status)
			}
		})
	}
}
