package domain

import (
	"time"

	"github.com/google/uuid"
)

type HousekeepingStatus string

const (
	StatusDirty      HousekeepingStatus = "dirty"
	StatusCleaning   HousekeepingStatus = "cleaning"
	StatusInspection HousekeepingStatus = "inspection"
	StatusReady      HousekeepingStatus = "ready"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

type Task struct {
	ID               uuid.UUID          `json:"id"`
	TenantID         uuid.UUID          `json:"tenant_id"`
	PropertyID       uuid.UUID          `json:"property_id"`
	UnitID           uuid.UUID          `json:"unit_id"`
	AssignedTo       *uuid.UUID         `json:"assigned_to,omitempty"`
	Status           HousekeepingStatus `json:"status"`
	Priority         Priority           `json:"priority"`
	TaskType         string             `json:"task_type"`
	Notes            string             `json:"notes,omitempty"`
	StartedAt        *time.Time         `json:"started_at,omitempty"`
	CompletedAt      *time.Time         `json:"completed_at,omitempty"`
	InspectedBy      *uuid.UUID         `json:"inspected_by,omitempty"`
	InspectedAt      *time.Time         `json:"inspected_at,omitempty"`
	EstimatedMinutes int                `json:"estimated_minutes"`
	ActualMinutes    *int               `json:"actual_minutes,omitempty"`
	CreatedBy        uuid.UUID          `json:"created_by"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
	// Joined fields
	UnitNumber   string `json:"unit_number,omitempty"`
	AssigneeName string `json:"assignee_name,omitempty"`
}

type CreateTaskInput struct {
	PropertyID       uuid.UUID  `json:"property_id" validate:"required"`
	UnitID           uuid.UUID  `json:"unit_id" validate:"required"`
	AssignedTo       *uuid.UUID `json:"assigned_to"`
	Priority         string     `json:"priority" validate:"required,oneof=low normal high urgent"`
	TaskType         string     `json:"task_type" validate:"required,oneof=checkout_clean stay_clean deep_clean maintenance"`
	Notes            string     `json:"notes" validate:"omitempty,max=1000"`
	EstimatedMinutes int        `json:"estimated_minutes" validate:"omitempty,min=5,max=480"`
}

type AssignTaskInput struct {
	TaskID     uuid.UUID `json:"task_id" validate:"required"`
	AssignedTo uuid.UUID `json:"assigned_to" validate:"required"`
}

type UpdateStatusInput struct {
	TaskID        uuid.UUID `json:"task_id" validate:"required"`
	Status        string    `json:"status" validate:"required,oneof=dirty cleaning inspection ready"`
	ActualMinutes *int      `json:"actual_minutes" validate:"omitempty,min=1"`
	Notes         string    `json:"notes" validate:"omitempty,max=1000"`
}

type TaskFilter struct {
	PropertyID *uuid.UUID
	Status     string
	AssignedTo *uuid.UUID
	Priority   string
}
