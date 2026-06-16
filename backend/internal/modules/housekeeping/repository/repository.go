package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stayflow/stayflow-track/internal/modules/housekeeping/domain"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *domain.Task) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO housekeeping_tasks (tenant_id, property_id, unit_id, assigned_to, status, priority, task_type, notes, estimated_minutes, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING id, created_at, updated_at`,
		task.TenantID, task.PropertyID, task.UnitID, task.AssignedTo,
		task.Status, task.Priority, task.TaskType, task.Notes,
		task.EstimatedMinutes, task.CreatedBy,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create task: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Task, error) {
	var t domain.Task
	err := r.pool.QueryRow(ctx,
		`SELECT t.id, t.tenant_id, t.property_id, t.unit_id, t.assigned_to,
		        t.status, t.priority, t.task_type, COALESCE(t.notes, ''),
		        t.started_at, t.completed_at, t.inspected_by, t.inspected_at,
		        COALESCE(t.estimated_minutes, 0), t.actual_minutes, t.created_by, t.created_at, t.updated_at,
		        u.unit_number,
		        COALESCE(usr.first_name || ' ' || usr.last_name, '')
		 FROM housekeeping_tasks t
		 JOIN units u ON t.unit_id = u.id
		 LEFT JOIN users usr ON t.assigned_to = usr.id
		 WHERE t.id = $1 AND t.tenant_id = $2`,
		id, tenantID,
	).Scan(&t.ID, &t.TenantID, &t.PropertyID, &t.UnitID, &t.AssignedTo,
		&t.Status, &t.Priority, &t.TaskType, &t.Notes,
		&t.StartedAt, &t.CompletedAt, &t.InspectedBy, &t.InspectedAt,
		&t.EstimatedMinutes, &t.ActualMinutes, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
		&t.UnitNumber, &t.AssigneeName)
	if err == pgx.ErrNoRows {
		return nil, apperrors.NotFound("housekeeping_task", id.String())
	}
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	return &t, nil
}

func (r *Repository) List(ctx context.Context, tenantID uuid.UUID, filter domain.TaskFilter, limit, offset int) ([]domain.Task, int64, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("t.tenant_id = $%d", argIdx))
	args = append(args, tenantID)
	argIdx++

	if filter.PropertyID != nil {
		conditions = append(conditions, fmt.Sprintf("t.property_id = $%d", argIdx))
		args = append(args, *filter.PropertyID)
		argIdx++
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("t.status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.AssignedTo != nil {
		conditions = append(conditions, fmt.Sprintf("t.assigned_to = $%d", argIdx))
		args = append(args, *filter.AssignedTo)
		argIdx++
	}
	if filter.Priority != "" {
		conditions = append(conditions, fmt.Sprintf("t.priority = $%d", argIdx))
		args = append(args, filter.Priority)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	var count int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM housekeeping_tasks t WHERE %s`, where)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&count); err != nil {
		return nil, 0, fmt.Errorf("count tasks: %w", err)
	}

	query := fmt.Sprintf(
		`SELECT t.id, t.tenant_id, t.property_id, t.unit_id, t.assigned_to,
		        t.status, t.priority, t.task_type, COALESCE(t.notes, ''),
		        t.started_at, t.completed_at, t.inspected_by, t.inspected_at,
		        COALESCE(t.estimated_minutes, 0), t.actual_minutes, t.created_by, t.created_at, t.updated_at,
		        u.unit_number,
		        COALESCE(usr.first_name || ' ' || usr.last_name, '')
		 FROM housekeeping_tasks t
		 JOIN units u ON t.unit_id = u.id
		 LEFT JOIN users usr ON t.assigned_to = usr.id
		 WHERE %s
		 ORDER BY CASE t.priority WHEN 'urgent' THEN 0 WHEN 'high' THEN 1 WHEN 'normal' THEN 2 ELSE 3 END, t.created_at DESC
		 LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.TenantID, &t.PropertyID, &t.UnitID, &t.AssignedTo,
			&t.Status, &t.Priority, &t.TaskType, &t.Notes,
			&t.StartedAt, &t.CompletedAt, &t.InspectedBy, &t.InspectedAt,
			&t.EstimatedMinutes, &t.ActualMinutes, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
			&t.UnitNumber, &t.AssigneeName); err != nil {
			return nil, 0, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	if tasks == nil {
		tasks = []domain.Task{}
	}
	return tasks, count, nil
}

func (r *Repository) Assign(ctx context.Context, taskID, tenantID, assignedTo uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE housekeeping_tasks SET assigned_to = $3 WHERE id = $1 AND tenant_id = $2`,
		taskID, tenantID, assignedTo,
	)
	if err != nil {
		return fmt.Errorf("assign task: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.NotFound("housekeeping_task", taskID.String())
	}
	return nil
}

func (r *Repository) UpdateStatus(ctx context.Context, taskID, tenantID, userID uuid.UUID, status string, actualMinutes *int, notes string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	now := time.Now()

	switch status {
	case "cleaning":
		_, err = tx.Exec(ctx,
			`UPDATE housekeeping_tasks SET status = $3, started_at = $4 WHERE id = $1 AND tenant_id = $2`,
			taskID, tenantID, status, now)
	case "inspection":
		_, err = tx.Exec(ctx,
			`UPDATE housekeeping_tasks SET status = $3, completed_at = $4, actual_minutes = $5
			 WHERE id = $1 AND tenant_id = $2`,
			taskID, tenantID, status, now, actualMinutes)
	case "ready":
		_, err = tx.Exec(ctx,
			`UPDATE housekeeping_tasks SET status = $3, inspected_by = $4, inspected_at = $5
			 WHERE id = $1 AND tenant_id = $2`,
			taskID, tenantID, status, userID, now)
		if err == nil {
			// Also update unit status to available
			_, err = tx.Exec(ctx,
				`UPDATE units SET status = 'available'
				 WHERE id = (SELECT unit_id FROM housekeeping_tasks WHERE id = $1) AND tenant_id = $2`,
				taskID, tenantID)
		}
	default:
		_, err = tx.Exec(ctx,
			`UPDATE housekeeping_tasks SET status = $3 WHERE id = $1 AND tenant_id = $2`,
			taskID, tenantID, status)
	}

	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	if notes != "" {
		_, _ = tx.Exec(ctx,
			`UPDATE housekeeping_tasks SET notes = COALESCE(notes, '') || E'\n' || $3 WHERE id = $1 AND tenant_id = $2`,
			taskID, tenantID, notes)
	}

	return tx.Commit(ctx)
}

// GetStats returns status counts for a property.
func (r *Repository) GetStats(ctx context.Context, tenantID, propertyID uuid.UUID) (map[string]int, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT status::TEXT, COUNT(*)::INT FROM housekeeping_tasks
		 WHERE tenant_id = $1 AND property_id = $2 AND completed_at IS NULL
		 GROUP BY status`,
		tenantID, propertyID,
	)
	if err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}
	defer rows.Close()

	stats := map[string]int{"dirty": 0, "cleaning": 0, "inspection": 0, "ready": 0}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats[status] = count
	}
	return stats, nil
}
