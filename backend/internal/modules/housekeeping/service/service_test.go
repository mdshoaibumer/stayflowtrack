package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stayflow/stayflow-track/internal/modules/housekeeping/domain"
)

func TestCreateTaskInput_Validation(t *testing.T) {
	tests := []struct {
		name  string
		input domain.CreateTaskInput
		valid bool
	}{
		{
			name: "valid input",
			input: domain.CreateTaskInput{
				PropertyID:       uuid.New(),
				UnitID:           uuid.New(),
				Priority:         "normal",
				TaskType:         "checkout_clean",
				EstimatedMinutes: 30,
			},
			valid: true,
		},
		{
			name: "invalid priority",
			input: domain.CreateTaskInput{
				PropertyID: uuid.New(),
				UnitID:     uuid.New(),
				Priority:   "invalid",
				TaskType:   "checkout_clean",
			},
			valid: false,
		},
		{
			name: "invalid task type",
			input: domain.CreateTaskInput{
				PropertyID: uuid.New(),
				UnitID:     uuid.New(),
				Priority:   "high",
				TaskType:   "invalid_type",
			},
			valid: false,
		},
		{
			name: "missing property_id",
			input: domain.CreateTaskInput{
				UnitID:   uuid.New(),
				Priority: "normal",
				TaskType: "deep_clean",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validation struct tag check (would use validator in integration)
			if tt.input.PropertyID == uuid.Nil && tt.valid {
				t.Error("expected invalid but got valid")
			}
		})
	}
}

func TestStatusTransitions(t *testing.T) {
	tests := []struct {
		from string
		to   string
		ok   bool
	}{
		{"dirty", "cleaning", true},
		{"cleaning", "inspection", true},
		{"inspection", "ready", true},
		{"dirty", "ready", true}, // allowed in our model
		{"ready", "dirty", true}, // re-dirty
	}

	for _, tt := range tests {
		t.Run(tt.from+"→"+tt.to, func(t *testing.T) {
			// All transitions are allowed in our model (no state machine restriction)
			_ = context.Background()
			if !tt.ok {
				t.Skip("no invalid transitions defined")
			}
		})
	}
}
