package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/stayflow/stayflow-track/internal/modules/housekeeping/domain"
	"github.com/stayflow/stayflow-track/internal/modules/housekeeping/repository"
)

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTask(ctx context.Context, tenantID, userID uuid.UUID, input domain.CreateTaskInput) (*domain.Task, error) {
	estMin := input.EstimatedMinutes
	if estMin == 0 {
		estMin = 30
	}

	task := &domain.Task{
		TenantID:         tenantID,
		PropertyID:       input.PropertyID,
		UnitID:           input.UnitID,
		AssignedTo:       input.AssignedTo,
		Status:           domain.StatusDirty,
		Priority:         domain.Priority(input.Priority),
		TaskType:         input.TaskType,
		Notes:            input.Notes,
		EstimatedMinutes: estMin,
		CreatedBy:        userID,
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *Service) GetTask(ctx context.Context, taskID, tenantID uuid.UUID) (*domain.Task, error) {
	return s.repo.GetByID(ctx, taskID, tenantID)
}

func (s *Service) ListTasks(ctx context.Context, tenantID uuid.UUID, filter domain.TaskFilter, limit, offset int) ([]domain.Task, int64, error) {
	return s.repo.List(ctx, tenantID, filter, limit, offset)
}

func (s *Service) AssignTask(ctx context.Context, tenantID uuid.UUID, input domain.AssignTaskInput) error {
	return s.repo.Assign(ctx, input.TaskID, tenantID, input.AssignedTo)
}

func (s *Service) UpdateStatus(ctx context.Context, tenantID, userID uuid.UUID, input domain.UpdateStatusInput) error {
	return s.repo.UpdateStatus(ctx, input.TaskID, tenantID, userID, input.Status, input.ActualMinutes, input.Notes)
}

func (s *Service) GetStats(ctx context.Context, tenantID, propertyID uuid.UUID) (map[string]int, error) {
	return s.repo.GetStats(ctx, tenantID, propertyID)
}
