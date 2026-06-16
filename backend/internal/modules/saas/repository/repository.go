package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stayflow/stayflow-track/internal/modules/saas/domain"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Plans

func (r *Repository) ListPlans(ctx context.Context, activeOnly bool) ([]domain.SubscriptionPlan, error) {
	query := `SELECT id, name, slug, COALESCE(description, ''), tier, price_monthly, price_yearly, currency,
	           max_properties, max_units, max_users, features, is_active, trial_days, sort_order, created_at, updated_at
	          FROM subscription_plans`
	if activeOnly {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY sort_order`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list plans: %w", err)
	}
	defer rows.Close()

	var plans []domain.SubscriptionPlan
	for rows.Next() {
		var p domain.SubscriptionPlan
		var featuresJSON []byte
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.Tier,
			&p.PriceMonthly, &p.PriceYearly, &p.Currency,
			&p.MaxProperties, &p.MaxUnits, &p.MaxUsers,
			&featuresJSON, &p.IsActive, &p.TrialDays, &p.SortOrder,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan plan: %w", err)
		}
		_ = json.Unmarshal(featuresJSON, &p.Features)
		plans = append(plans, p)
	}
	if plans == nil {
		plans = []domain.SubscriptionPlan{}
	}
	return plans, nil
}

func (r *Repository) GetPlanBySlug(ctx context.Context, slug string) (*domain.SubscriptionPlan, error) {
	var p domain.SubscriptionPlan
	var featuresJSON []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, slug, COALESCE(description, ''), tier, price_monthly, price_yearly, currency,
		        max_properties, max_units, max_users, features, is_active, trial_days, sort_order, created_at, updated_at
		 FROM subscription_plans WHERE slug = $1`, slug,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.Tier,
		&p.PriceMonthly, &p.PriceYearly, &p.Currency,
		&p.MaxProperties, &p.MaxUnits, &p.MaxUsers,
		&featuresJSON, &p.IsActive, &p.TrialDays, &p.SortOrder,
		&p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, apperrors.NotFound("plan", slug)
	}
	if err != nil {
		return nil, fmt.Errorf("get plan: %w", err)
	}
	_ = json.Unmarshal(featuresJSON, &p.Features)
	return &p, nil
}

// Subscriptions

func (r *Repository) GetSubscription(ctx context.Context, tenantID uuid.UUID) (*domain.TenantSubscription, error) {
	var s domain.TenantSubscription
	err := r.pool.QueryRow(ctx,
		`SELECT ts.id, ts.tenant_id, ts.plan_id, ts.status, ts.billing_cycle,
		        ts.current_period_start, ts.current_period_end, ts.trial_ends_at, ts.cancelled_at,
		        COALESCE(ts.razorpay_subscription_id, ''), COALESCE(ts.razorpay_customer_id, ''), ts.created_at, ts.updated_at,
		        sp.name, sp.tier
		 FROM tenant_subscriptions ts
		 JOIN subscription_plans sp ON ts.plan_id = sp.id
		 WHERE ts.tenant_id = $1`, tenantID,
	).Scan(&s.ID, &s.TenantID, &s.PlanID, &s.Status, &s.BillingCycle,
		&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.TrialEndsAt, &s.CancelledAt,
		&s.RazorpaySubscriptionID, &s.RazorpayCustomerID, &s.CreatedAt, &s.UpdatedAt,
		&s.PlanName, &s.PlanTier)
	if err == pgx.ErrNoRows {
		return nil, apperrors.NotFound("subscription", tenantID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("get subscription: %w", err)
	}
	return &s, nil
}

func (r *Repository) CreateSubscription(ctx context.Context, sub *domain.TenantSubscription) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO tenant_subscriptions (tenant_id, plan_id, status, billing_cycle, current_period_start, current_period_end, trial_ends_at, razorpay_subscription_id, razorpay_customer_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, created_at, updated_at`,
		sub.TenantID, sub.PlanID, sub.Status, sub.BillingCycle,
		sub.CurrentPeriodStart, sub.CurrentPeriodEnd, sub.TrialEndsAt,
		sub.RazorpaySubscriptionID, sub.RazorpayCustomerID,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)
}

func (r *Repository) UpdateSubscription(ctx context.Context, sub *domain.TenantSubscription) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE tenant_subscriptions SET plan_id = $2, status = $3, billing_cycle = $4,
		        current_period_start = $5, current_period_end = $6, cancelled_at = $7,
		        razorpay_subscription_id = $8, razorpay_customer_id = $9
		 WHERE id = $1`,
		sub.ID, sub.PlanID, sub.Status, sub.BillingCycle,
		sub.CurrentPeriodStart, sub.CurrentPeriodEnd, sub.CancelledAt,
		sub.RazorpaySubscriptionID, sub.RazorpayCustomerID,
	)
	return err
}

func (r *Repository) ActivateSubscription(ctx context.Context, tenantID uuid.UUID, periodEnd time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE tenant_subscriptions SET status = 'active', current_period_end = $2 WHERE tenant_id = $1`,
		tenantID, periodEnd,
	)
	return err
}

func (r *Repository) CancelSubscription(ctx context.Context, tenantID uuid.UUID) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE tenant_subscriptions SET status = 'cancelled', cancelled_at = $2 WHERE tenant_id = $1`,
		tenantID, now,
	)
	return err
}

// Billing Events

func (r *Repository) CreateBillingEvent(ctx context.Context, event *domain.BillingEvent) error {
	metadataJSON, _ := json.Marshal(event.Metadata)
	return r.pool.QueryRow(ctx,
		`INSERT INTO billing_events (tenant_id, subscription_id, event_type, amount, currency, razorpay_payment_id, razorpay_order_id, razorpay_signature, status, metadata)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING id, created_at`,
		event.TenantID, event.SubscriptionID, event.EventType, event.Amount, event.Currency,
		event.RazorpayPaymentID, event.RazorpayOrderID, event.RazorpaySignature, event.Status, metadataJSON,
	).Scan(&event.ID, &event.CreatedAt)
}

func (r *Repository) ListBillingEvents(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.BillingEvent, int64, error) {
	var count int64
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM billing_events WHERE tenant_id = $1`, tenantID).Scan(&count)

	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, subscription_id, event_type, amount, currency,
		        COALESCE(razorpay_payment_id, ''), COALESCE(razorpay_order_id, ''), status, metadata, created_at
		 FROM billing_events WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		tenantID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	var events []domain.BillingEvent
	for rows.Next() {
		var e domain.BillingEvent
		var metaJSON []byte
		if err := rows.Scan(&e.ID, &e.TenantID, &e.SubscriptionID, &e.EventType, &e.Amount, &e.Currency,
			&e.RazorpayPaymentID, &e.RazorpayOrderID, &e.Status, &metaJSON, &e.CreatedAt); err != nil {
			return nil, 0, err
		}
		if metaJSON != nil {
			_ = json.Unmarshal(metaJSON, &e.Metadata)
		}
		events = append(events, e)
	}
	if events == nil {
		events = []domain.BillingEvent{}
	}
	return events, count, nil
}

// Onboarding

func (r *Repository) InitOnboarding(ctx context.Context, tenantID, propertyID uuid.UUID) error {
	steps := []struct {
		key   string
		name  string
		order int
	}{
		{"property_details", "Complete Property Details", 1},
		{"unit_types", "Create Unit Types", 2},
		{"units", "Add Units/Rooms", 3},
		{"rates", "Set Up Rates", 4},
		{"staff", "Add Staff Users", 5},
		{"guest_import", "Import Existing Guests", 6},
		{"payment_setup", "Configure Payment Methods", 7},
		{"go_live", "Go Live", 8},
	}

	for _, step := range steps {
		_, err := r.pool.Exec(ctx,
			`INSERT INTO onboarding_checklists (tenant_id, property_id, step_key, step_name, sort_order)
			 VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`,
			tenantID, propertyID, step.key, step.name, step.order,
		)
		if err != nil {
			return fmt.Errorf("init onboarding step %s: %w", step.key, err)
		}
	}
	return nil
}

func (r *Repository) GetOnboardingSteps(ctx context.Context, tenantID, propertyID uuid.UUID) ([]domain.OnboardingStep, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, property_id, step_key, step_name, is_completed, completed_at, completed_by, sort_order, created_at
		 FROM onboarding_checklists WHERE tenant_id = $1 AND property_id = $2 ORDER BY sort_order`,
		tenantID, propertyID,
	)
	if err != nil {
		return nil, fmt.Errorf("get onboarding: %w", err)
	}
	defer rows.Close()

	var steps []domain.OnboardingStep
	for rows.Next() {
		var s domain.OnboardingStep
		if err := rows.Scan(&s.ID, &s.TenantID, &s.PropertyID, &s.StepKey, &s.StepName,
			&s.IsCompleted, &s.CompletedAt, &s.CompletedBy, &s.SortOrder, &s.CreatedAt); err != nil {
			return nil, err
		}
		steps = append(steps, s)
	}
	if steps == nil {
		steps = []domain.OnboardingStep{}
	}
	return steps, nil
}

func (r *Repository) CompleteStep(ctx context.Context, tenantID, propertyID uuid.UUID, stepKey string, userID uuid.UUID) error {
	now := time.Now()
	result, err := r.pool.Exec(ctx,
		`UPDATE onboarding_checklists SET is_completed = true, completed_at = $4, completed_by = $5
		 WHERE tenant_id = $1 AND property_id = $2 AND step_key = $3 AND is_completed = false`,
		tenantID, propertyID, stepKey, now, userID,
	)
	if err != nil {
		return fmt.Errorf("complete step: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.NotFound("onboarding_step", stepKey)
	}
	return nil
}

// Admin

func (r *Repository) ListTenants(ctx context.Context, status string, limit, offset int) ([]domain.TenantOverview, int64, error) {
	var count int64
	_ = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM tenants WHERE ($1::VARCHAR = '' OR status = $1)`, status,
	).Scan(&count)

	rows, err := r.pool.Query(ctx,
		`SELECT t.id, t.name, t.email, t.status,
		        COALESCE(sp.name, 'None'), COALESCE(sp.tier::TEXT, 'none'),
		        COALESCE(ts.status::TEXT, 'none'),
		        (SELECT COUNT(*) FROM properties WHERE tenant_id = t.id),
		        (SELECT COUNT(*) FROM users WHERE tenant_id = t.id),
		        (SELECT COUNT(*) FROM units WHERE tenant_id = t.id),
		        t.created_at,
		        (SELECT MAX(last_login_at) FROM users WHERE tenant_id = t.id)
		 FROM tenants t
		 LEFT JOIN tenant_subscriptions ts ON t.id = ts.tenant_id
		 LEFT JOIN subscription_plans sp ON ts.plan_id = sp.id
		 WHERE ($1::VARCHAR = '' OR t.status = $1)
		 ORDER BY t.created_at DESC LIMIT $2 OFFSET $3`,
		status, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []domain.TenantOverview
	for rows.Next() {
		var t domain.TenantOverview
		if err := rows.Scan(&t.TenantID, &t.TenantName, &t.TenantEmail, &t.TenantStatus,
			&t.PlanName, &t.PlanTier, &t.SubscriptionStatus,
			&t.PropertyCount, &t.UserCount, &t.UnitCount,
			&t.CreatedAt, &t.LastLoginAt); err != nil {
			return nil, 0, err
		}
		tenants = append(tenants, t)
	}
	if tenants == nil {
		tenants = []domain.TenantOverview{}
	}
	return tenants, count, nil
}

func (r *Repository) GetSaaSMetrics(ctx context.Context) (*domain.SaaSMetrics, error) {
	var m domain.SaaSMetrics
	err := r.pool.QueryRow(ctx,
		`SELECT
			(SELECT COUNT(*) FROM tenants),
			(SELECT COUNT(*) FROM tenants WHERE status = 'active'),
			(SELECT COUNT(*) FROM tenant_subscriptions WHERE status = 'trialing'),
			(SELECT COUNT(*) FROM tenant_subscriptions WHERE status = 'active'),
			COALESCE((SELECT SUM(CASE WHEN ts.billing_cycle = 'monthly' THEN sp.price_monthly ELSE sp.price_yearly/12 END)
			 FROM tenant_subscriptions ts JOIN subscription_plans sp ON ts.plan_id = sp.id WHERE ts.status = 'active'), 0),
			(SELECT COUNT(*) FROM properties),
			(SELECT COUNT(*) FROM units)`,
	).Scan(&m.TotalTenants, &m.ActiveTenants, &m.TrialingTenants, &m.PaidTenants,
		&m.MRR, &m.TotalProperties, &m.TotalUnits)
	if err != nil {
		return nil, fmt.Errorf("get saas metrics: %w", err)
	}
	m.ARR = m.MRR.Mul(decimal.NewFromInt(12))
	if m.PaidTenants > 0 {
		m.AvgRevenuePerTenant = m.MRR.Div(decimal.NewFromInt(int64(m.PaidTenants)))
	}
	return &m, nil
}

// Usage check for plan limits
func (r *Repository) GetUsage(ctx context.Context, tenantID uuid.UUID) (properties, units, users int, err error) {
	err = r.pool.QueryRow(ctx,
		`SELECT
			(SELECT COUNT(*) FROM properties WHERE tenant_id = $1),
			(SELECT COUNT(*) FROM units WHERE tenant_id = $1),
			(SELECT COUNT(*) FROM users WHERE tenant_id = $1)`,
		tenantID,
	).Scan(&properties, &units, &users)
	return
}
