package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stayflow/stayflow-track/internal/modules/saas/domain"
	"github.com/stayflow/stayflow-track/internal/modules/saas/service"
)

// MockRepo implements service.Repository for testing.
type MockRepo struct {
	plans         []domain.SubscriptionPlan
	subscription  *domain.TenantSubscription
	billingEvents []domain.BillingEvent
	onboarding    []domain.OnboardingStep
	tenants       []domain.TenantOverview
	metrics       *domain.SaaSMetrics
	usage         [3]int // props, units, users

	createdSub    *domain.TenantSubscription
	cancelledID   uuid.UUID
	completedStep string
}

func (m *MockRepo) ListPlans(ctx context.Context, activeOnly bool) ([]domain.SubscriptionPlan, error) {
	if activeOnly {
		var active []domain.SubscriptionPlan
		for _, p := range m.plans {
			if p.IsActive {
				active = append(active, p)
			}
		}
		return active, nil
	}
	return m.plans, nil
}

func (m *MockRepo) GetPlanBySlug(ctx context.Context, slug string) (*domain.SubscriptionPlan, error) {
	for _, p := range m.plans {
		if p.Slug == slug {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("plan not found")
}

func (m *MockRepo) GetSubscription(ctx context.Context, tenantID uuid.UUID) (*domain.TenantSubscription, error) {
	if m.subscription != nil {
		return m.subscription, nil
	}
	return nil, fmt.Errorf("subscription not found")
}

func (m *MockRepo) CreateSubscription(ctx context.Context, sub *domain.TenantSubscription) error {
	sub.ID = uuid.New()
	m.createdSub = sub
	return nil
}

func (m *MockRepo) UpdateSubscription(ctx context.Context, sub *domain.TenantSubscription) error {
	m.subscription = sub
	return nil
}

func (m *MockRepo) ActivateSubscription(ctx context.Context, tenantID uuid.UUID, periodEnd time.Time) error {
	return nil
}

func (m *MockRepo) CancelSubscription(ctx context.Context, tenantID uuid.UUID) error {
	m.cancelledID = tenantID
	return nil
}

func (m *MockRepo) CreateBillingEvent(ctx context.Context, event *domain.BillingEvent) error {
	event.ID = uuid.New()
	m.billingEvents = append(m.billingEvents, *event)
	return nil
}

func (m *MockRepo) ListBillingEvents(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.BillingEvent, int64, error) {
	return m.billingEvents, int64(len(m.billingEvents)), nil
}

func (m *MockRepo) InitOnboarding(ctx context.Context, tenantID, propertyID uuid.UUID) error {
	return nil
}

func (m *MockRepo) GetOnboardingSteps(ctx context.Context, tenantID, propertyID uuid.UUID) ([]domain.OnboardingStep, error) {
	return m.onboarding, nil
}

func (m *MockRepo) CompleteStep(ctx context.Context, tenantID, propertyID uuid.UUID, stepKey string, userID uuid.UUID) error {
	m.completedStep = stepKey
	return nil
}

func (m *MockRepo) ListTenants(ctx context.Context, status string, limit, offset int) ([]domain.TenantOverview, int64, error) {
	return m.tenants, int64(len(m.tenants)), nil
}

func (m *MockRepo) GetSaaSMetrics(ctx context.Context) (*domain.SaaSMetrics, error) {
	return m.metrics, nil
}

func (m *MockRepo) GetUsage(ctx context.Context, tenantID uuid.UUID) (int, int, int, error) {
	return m.usage[0], m.usage[1], m.usage[2], nil
}

// Tests

func TestCreateSubscription_Trial(t *testing.T) {
	planID := uuid.New()
	repo := &MockRepo{
		plans: []domain.SubscriptionPlan{
			{ID: planID, Slug: "starter", Name: "Starter", IsActive: true, TrialDays: 14, PriceMonthly: decimal.NewFromInt(999)},
		},
	}

	svc := service.New(repo, nil)

	tenantID := uuid.New()
	sub, err := svc.CreateSubscription(context.Background(), domain.CreateSubscriptionInput{
		TenantID:     tenantID,
		PlanSlug:     "starter",
		BillingCycle: "monthly",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.Status != domain.SubTrialing {
		t.Errorf("expected status trialing, got %s", sub.Status)
	}
	if sub.TrialEndsAt == nil {
		t.Error("expected trial_ends_at to be set")
	}
	if sub.PlanID != planID {
		t.Errorf("expected plan ID %s, got %s", planID, sub.PlanID)
	}
}

func TestCreateSubscription_NoPlan(t *testing.T) {
	repo := &MockRepo{plans: []domain.SubscriptionPlan{}}
	svc := service.New(repo, nil)

	_, err := svc.CreateSubscription(context.Background(), domain.CreateSubscriptionInput{
		TenantID: uuid.New(),
		PlanSlug: "nonexistent",
	})

	if err == nil {
		t.Error("expected error for nonexistent plan")
	}
}

func TestChangePlan_ExceedsLimits(t *testing.T) {
	subID := uuid.New()
	tenantID := uuid.New()
	oldPlanID := uuid.New()
	newPlanID := uuid.New()

	repo := &MockRepo{
		plans: []domain.SubscriptionPlan{
			{ID: oldPlanID, Slug: "professional", MaxProperties: 10, MaxUnits: 200, MaxUsers: 20},
			{ID: newPlanID, Slug: "starter", MaxProperties: 2, MaxUnits: 30, MaxUsers: 5},
		},
		subscription: &domain.TenantSubscription{
			ID:       subID,
			TenantID: tenantID,
			PlanID:   oldPlanID,
			Status:   domain.SubActive,
			PlanTier: "professional",
		},
		usage: [3]int{5, 50, 10}, // exceeds starter limits
	}

	svc := service.New(repo, nil)

	_, err := svc.ChangePlan(context.Background(), domain.ChangePlanInput{
		TenantID: tenantID,
		PlanSlug: "starter",
	})

	if err == nil {
		t.Error("expected error when usage exceeds new plan limits")
	}
}

func TestCancelSubscription(t *testing.T) {
	tenantID := uuid.New()
	subID := uuid.New()

	repo := &MockRepo{
		subscription: &domain.TenantSubscription{
			ID:       subID,
			TenantID: tenantID,
			Status:   domain.SubActive,
		},
	}

	svc := service.New(repo, nil)

	err := svc.CancelSubscription(context.Background(), domain.CancelSubscriptionInput{
		TenantID: tenantID,
		Reason:   "too expensive",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.cancelledID != tenantID {
		t.Error("expected tenant to be cancelled")
	}
}

func TestCheckLimits_PropertyExceeded(t *testing.T) {
	tenantID := uuid.New()
	planID := uuid.New()

	repo := &MockRepo{
		plans: []domain.SubscriptionPlan{
			{ID: planID, Slug: "starter", MaxProperties: 2, MaxUnits: 30, MaxUsers: 5},
		},
		subscription: &domain.TenantSubscription{
			TenantID: tenantID,
			PlanID:   planID,
			PlanTier: "starter",
			Status:   domain.SubActive,
		},
		usage: [3]int{2, 10, 3}, // at property limit
	}

	svc := service.New(repo, nil)

	err := svc.CheckLimits(context.Background(), tenantID, "property")
	if err == nil {
		t.Error("expected error when at property limit")
	}
}

func TestCheckLimits_WithinLimits(t *testing.T) {
	tenantID := uuid.New()
	planID := uuid.New()

	repo := &MockRepo{
		plans: []domain.SubscriptionPlan{
			{ID: planID, Slug: "starter", MaxProperties: 2, MaxUnits: 30, MaxUsers: 5},
		},
		subscription: &domain.TenantSubscription{
			TenantID: tenantID,
			PlanID:   planID,
			PlanTier: "starter",
			Status:   domain.SubActive,
		},
		usage: [3]int{1, 10, 3}, // within limits
	}

	svc := service.New(repo, nil)

	err := svc.CheckLimits(context.Background(), tenantID, "property")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
