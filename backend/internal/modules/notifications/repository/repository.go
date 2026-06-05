package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stayflow/stayflow-track/internal/modules/notifications/domain"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetTemplate(ctx context.Context, tenantID uuid.UUID, eventType string, channel domain.Channel) (*domain.Template, error) {
	var t domain.Template
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, event_type, channel, template_name, template_body, is_active, created_at, updated_at
		 FROM notification_templates
		 WHERE tenant_id = $1 AND event_type = $2 AND channel = $3 AND is_active = true`,
		tenantID, eventType, channel,
	).Scan(&t.ID, &t.TenantID, &t.EventType, &t.Channel, &t.TemplateName, &t.TemplateBody, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, apperrors.NotFound("notification_template", eventType)
	}
	if err != nil {
		return nil, fmt.Errorf("get template: %w", err)
	}
	return &t, nil
}

func (r *Repository) UpsertTemplate(ctx context.Context, t *domain.Template) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO notification_templates (tenant_id, event_type, channel, template_name, template_body, is_active)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (tenant_id, event_type, channel)
		 DO UPDATE SET template_name = $4, template_body = $5, is_active = $6
		 RETURNING id, created_at, updated_at`,
		t.TenantID, t.EventType, t.Channel, t.TemplateName, t.TemplateBody, t.IsActive,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert template: %w", err)
	}
	return nil
}

func (r *Repository) ListTemplates(ctx context.Context, tenantID uuid.UUID) ([]domain.Template, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, event_type, channel, template_name, template_body, is_active, created_at, updated_at
		 FROM notification_templates WHERE tenant_id = $1 ORDER BY event_type, channel`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	defer rows.Close()

	var templates []domain.Template
	for rows.Next() {
		var t domain.Template
		if err := rows.Scan(&t.ID, &t.TenantID, &t.EventType, &t.Channel, &t.TemplateName, &t.TemplateBody, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	if templates == nil {
		templates = []domain.Template{}
	}
	return templates, nil
}

func (r *Repository) CreateLog(ctx context.Context, log *domain.NotificationLog) error {
	payloadJSON, _ := json.Marshal(log.Payload)
	err := r.pool.QueryRow(ctx,
		`INSERT INTO notification_logs (tenant_id, template_id, event_type, channel, recipient_phone, recipient_email, status, provider, provider_message_id, payload, error_message, sent_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 RETURNING id, created_at`,
		log.TenantID, log.TemplateID, log.EventType, log.Channel,
		log.RecipientPhone, log.RecipientEmail, log.Status, log.Provider,
		log.ProviderMessageID, payloadJSON, log.ErrorMessage, log.SentAt,
	).Scan(&log.ID, &log.CreatedAt)
	if err != nil {
		return fmt.Errorf("create log: %w", err)
	}
	return nil
}

func (r *Repository) ListLogs(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.NotificationLog, int64, error) {
	var count int64
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notification_logs WHERE tenant_id = $1`, tenantID).Scan(&count)

	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, template_id, event_type, channel, recipient_phone, recipient_email,
		        status, provider, provider_message_id, payload, error_message, sent_at, delivered_at, created_at
		 FROM notification_logs WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		tenantID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list logs: %w", err)
	}
	defer rows.Close()

	var logs []domain.NotificationLog
	for rows.Next() {
		var l domain.NotificationLog
		var payloadJSON []byte
		if err := rows.Scan(&l.ID, &l.TenantID, &l.TemplateID, &l.EventType, &l.Channel,
			&l.RecipientPhone, &l.RecipientEmail, &l.Status, &l.Provider,
			&l.ProviderMessageID, &payloadJSON, &l.ErrorMessage, &l.SentAt, &l.DeliveredAt, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		if payloadJSON != nil {
			json.Unmarshal(payloadJSON, &l.Payload)
		}
		logs = append(logs, l)
	}
	if logs == nil {
		logs = []domain.NotificationLog{}
	}
	return logs, count, nil
}
