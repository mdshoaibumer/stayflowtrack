package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stayflow/stayflow-track/internal/modules/saas/domain"
	"github.com/stayflow/stayflow-track/internal/modules/saas/razorpay"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

// Repository defines the data access interface for SaaS management.
type Repository interface {
	ListPlans(ctx context.Context, activeOnly bool) ([]domain.SubscriptionPlan, error)
	GetPlanBySlug(ctx context.Context, slug string) (*domain.SubscriptionPlan, error)
	GetSubscription(ctx context.Context, tenantID uuid.UUID) (*domain.TenantSubscription, error)
	CreateSubscription(ctx context.Context, sub *domain.TenantSubscription) error
	UpdateSubscription(ctx context.Context, sub *domain.TenantSubscription) error
	ActivateSubscription(ctx context.Context, tenantID uuid.UUID, periodEnd time.Time) error
	CancelSubscription(ctx context.Context, tenantID uuid.UUID) error
	CreateBillingEvent(ctx context.Context, event *domain.BillingEvent) error
	ListBillingEvents(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.BillingEvent, int64, error)
	InitOnboarding(ctx context.Context, tenantID, propertyID uuid.UUID) error
	GetOnboardingSteps(ctx context.Context, tenantID, propertyID uuid.UUID) ([]domain.OnboardingStep, error)
	CompleteStep(ctx context.Context, tenantID, propertyID uuid.UUID, stepKey string, userID uuid.UUID) error
	ListTenants(ctx context.Context, status string, limit, offset int) ([]domain.TenantOverview, int64, error)
	GetSaaSMetrics(ctx context.Context) (*domain.SaaSMetrics, error)
	GetUsage(ctx context.Context, tenantID uuid.UUID) (int, int, int, error)
}

type Service struct {
	repo     Repository
	razorpay *razorpay.Client
}

func New(repo Repository, rp *razorpay.Client) *Service {
	return &Service{repo: repo, razorpay: rp}
}

// Plans

func (s *Service) ListPlans(ctx context.Context, activeOnly bool) ([]domain.SubscriptionPlan, error) {
	return s.repo.ListPlans(ctx, activeOnly)
}

func (s *Service) GetPlan(ctx context.Context, slug string) (*domain.SubscriptionPlan, error) {
	return s.repo.GetPlanBySlug(ctx, slug)
}

// Subscriptions

func (s *Service) GetSubscription(ctx context.Context, tenantID uuid.UUID) (*domain.TenantSubscription, error) {
	return s.repo.GetSubscription(ctx, tenantID)
}

func (s *Service) CreateSubscription(ctx context.Context, input domain.CreateSubscriptionInput) (*domain.TenantSubscription, error) {
	plan, err := s.repo.GetPlanBySlug(ctx, input.PlanSlug)
	if err != nil {
		return nil, err
	}
	if !plan.IsActive {
		return nil, apperrors.BadRequest("plan is not active")
	}

	now := time.Now()
	var periodEnd time.Time
	var status domain.SubscriptionStatus
	var trialEnd *time.Time

	if plan.TrialDays > 0 {
		status = domain.SubTrialing
		te := now.AddDate(0, 0, plan.TrialDays)
		trialEnd = &te
		periodEnd = te
	} else {
		status = domain.SubActive
		if input.BillingCycle == "yearly" {
			periodEnd = now.AddDate(1, 0, 0)
		} else {
			periodEnd = now.AddDate(0, 1, 0)
		}
	}

	sub := &domain.TenantSubscription{
		TenantID:           input.TenantID,
		PlanID:             plan.ID,
		Status:             status,
		BillingCycle:       domain.BillingCycle(input.BillingCycle),
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   periodEnd,
		TrialEndsAt:        trialEnd,
	}

	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		return nil, fmt.Errorf("create subscription: %w", err)
	}

	// Log billing event
	s.repo.CreateBillingEvent(ctx, &domain.BillingEvent{
		TenantID:       input.TenantID,
		SubscriptionID: &sub.ID,
		EventType:      "subscription_created",
		Amount:         decimal.Zero,
		Currency:       plan.Currency,
		Status:         "success",
	})

	return sub, nil
}

func (s *Service) ChangePlan(ctx context.Context, input domain.ChangePlanInput) (*domain.TenantSubscription, error) {
	sub, err := s.repo.GetSubscription(ctx, input.TenantID)
	if err != nil {
		return nil, err
	}

	newPlan, err := s.repo.GetPlanBySlug(ctx, input.PlanSlug)
	if err != nil {
		return nil, err
	}

	// Check usage against new plan limits
	props, units, users, err := s.repo.GetUsage(ctx, input.TenantID)
	if err != nil {
		return nil, err
	}
	if props > newPlan.MaxProperties {
		return nil, apperrors.BadRequest(fmt.Sprintf("current properties (%d) exceed new plan limit (%d)", props, newPlan.MaxProperties))
	}
	if units > newPlan.MaxUnits {
		return nil, apperrors.BadRequest(fmt.Sprintf("current units (%d) exceed new plan limit (%d)", units, newPlan.MaxUnits))
	}
	if users > newPlan.MaxUsers {
		return nil, apperrors.BadRequest(fmt.Sprintf("current users (%d) exceed new plan limit (%d)", users, newPlan.MaxUsers))
	}

	sub.PlanID = newPlan.ID
	if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
		return nil, err
	}

	s.repo.CreateBillingEvent(ctx, &domain.BillingEvent{
		TenantID:       input.TenantID,
		SubscriptionID: &sub.ID,
		EventType:      "plan_changed",
		Status:         "success",
		Metadata:       map[string]interface{}{"new_plan": newPlan.Slug},
	})

	return sub, nil
}

func (s *Service) CancelSubscription(ctx context.Context, input domain.CancelSubscriptionInput) error {
	sub, err := s.repo.GetSubscription(ctx, input.TenantID)
	if err != nil {
		return err
	}

	// Cancel in Razorpay if exists
	if sub.RazorpaySubscriptionID != "" && s.razorpay != nil {
		_ = s.razorpay.CancelSubscription(ctx, sub.RazorpaySubscriptionID, true)
	}

	if err := s.repo.CancelSubscription(ctx, input.TenantID); err != nil {
		return err
	}

	s.repo.CreateBillingEvent(ctx, &domain.BillingEvent{
		TenantID:       input.TenantID,
		SubscriptionID: &sub.ID,
		EventType:      "subscription_cancelled",
		Status:         "success",
		Metadata:       map[string]interface{}{"reason": input.Reason},
	})

	return nil
}

// Razorpay Integration

type CreateCheckoutInput struct {
	TenantID     uuid.UUID `json:"tenant_id" validate:"required"`
	TenantName   string    `json:"tenant_name" validate:"required"`
	TenantEmail  string    `json:"tenant_email" validate:"required,email"`
	TenantPhone  string    `json:"tenant_phone"`
	PlanSlug     string    `json:"plan_slug" validate:"required"`
	BillingCycle string    `json:"billing_cycle" validate:"required,oneof=monthly yearly"`
}

type CheckoutResponse struct {
	OrderID  string `json:"order_id"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	KeyID    string `json:"key_id"`
	PlanName string `json:"plan_name"`
}

func (s *Service) CreateCheckout(ctx context.Context, input CreateCheckoutInput) (*CheckoutResponse, error) {
	plan, err := s.repo.GetPlanBySlug(ctx, input.PlanSlug)
	if err != nil {
		return nil, err
	}

	var price decimal.Decimal
	if input.BillingCycle == "yearly" {
		price = plan.PriceYearly
	} else {
		price = plan.PriceMonthly
	}

	if !price.IsPositive() {
		return nil, apperrors.BadRequest("free plan does not require payment")
	}

	amountPaise := price.Mul(decimal.NewFromInt(100)).IntPart()
	receipt := fmt.Sprintf("sub_%s_%s", input.TenantID.String()[:8], plan.Slug)

	if s.razorpay == nil {
		// Dev mode: return mock
		return &CheckoutResponse{
			OrderID:  "order_mock_" + uuid.New().String()[:8],
			Amount:   amountPaise,
			Currency: plan.Currency,
			KeyID:    "rzp_test_mock",
			PlanName: plan.Name,
		}, nil
	}

	order, err := s.razorpay.CreateOrder(ctx, amountPaise, plan.Currency, receipt)
	if err != nil {
		return nil, fmt.Errorf("razorpay order: %w", err)
	}

	// Log pending event
	s.repo.CreateBillingEvent(ctx, &domain.BillingEvent{
		TenantID:        input.TenantID,
		EventType:       "checkout_initiated",
		Amount:          price,
		Currency:        plan.Currency,
		RazorpayOrderID: order.ID,
		Status:          "pending",
	})

	return &CheckoutResponse{
		OrderID:  order.ID,
		Amount:   amountPaise,
		Currency: plan.Currency,
		KeyID:    s.razorpay.KeyID(),
		PlanName: plan.Name,
	}, nil
}

type VerifyPaymentInput struct {
	TenantID          uuid.UUID `json:"tenant_id" validate:"required"`
	RazorpayOrderID   string    `json:"razorpay_order_id" validate:"required"`
	RazorpayPaymentID string    `json:"razorpay_payment_id" validate:"required"`
	RazorpaySignature string    `json:"razorpay_signature" validate:"required"`
}

func (s *Service) VerifyPayment(ctx context.Context, input VerifyPaymentInput) error {
	// Verify signature
	if s.razorpay != nil {
		if !s.razorpay.VerifyPaymentSignature(input.RazorpayOrderID, input.RazorpayPaymentID, input.RazorpaySignature) {
			return apperrors.BadRequest("invalid payment signature")
		}
	}

	// Activate subscription
	sub, err := s.repo.GetSubscription(ctx, input.TenantID)
	if err != nil {
		return err
	}

	var periodEnd time.Time
	if sub.BillingCycle == domain.CycleYearly {
		periodEnd = time.Now().AddDate(1, 0, 0)
	} else {
		periodEnd = time.Now().AddDate(0, 1, 0)
	}

	if err := s.repo.ActivateSubscription(ctx, input.TenantID, periodEnd); err != nil {
		return err
	}

	// Log success event
	s.repo.CreateBillingEvent(ctx, &domain.BillingEvent{
		TenantID:          input.TenantID,
		SubscriptionID:    &sub.ID,
		EventType:         "payment_success",
		RazorpayPaymentID: input.RazorpayPaymentID,
		RazorpayOrderID:   input.RazorpayOrderID,
		RazorpaySignature: input.RazorpaySignature,
		Status:            "success",
	})

	return nil
}

// Billing History

func (s *Service) ListBillingEvents(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.BillingEvent, int64, error) {
	return s.repo.ListBillingEvents(ctx, tenantID, limit, offset)
}

// Onboarding

func (s *Service) InitOnboarding(ctx context.Context, input domain.OnboardPropertyInput) error {
	return s.repo.InitOnboarding(ctx, input.TenantID, input.PropertyID)
}

func (s *Service) GetOnboardingSteps(ctx context.Context, tenantID, propertyID uuid.UUID) ([]domain.OnboardingStep, error) {
	return s.repo.GetOnboardingSteps(ctx, tenantID, propertyID)
}

func (s *Service) CompleteStep(ctx context.Context, tenantID, propertyID uuid.UUID, stepKey string, userID uuid.UUID) error {
	return s.repo.CompleteStep(ctx, tenantID, propertyID, stepKey, userID)
}

// Admin

func (s *Service) ListTenants(ctx context.Context, status string, limit, offset int) ([]domain.TenantOverview, int64, error) {
	return s.repo.ListTenants(ctx, status, limit, offset)
}

func (s *Service) GetSaaSMetrics(ctx context.Context) (*domain.SaaSMetrics, error) {
	return s.repo.GetSaaSMetrics(ctx)
}

// CheckLimits verifies if tenant can add more resources.
func (s *Service) CheckLimits(ctx context.Context, tenantID uuid.UUID, resource string) error {
	sub, err := s.repo.GetSubscription(ctx, tenantID)
	if err != nil {
		// No subscription - allow (free tier behavior)
		return nil
	}

	plan, err := s.repo.GetPlanBySlug(ctx, sub.PlanTier)
	if err != nil {
		return nil
	}

	props, units, users, err := s.repo.GetUsage(ctx, tenantID)
	if err != nil {
		return err
	}

	switch resource {
	case "property":
		if props >= plan.MaxProperties {
			return apperrors.BadRequest(fmt.Sprintf("plan limit reached: max %d properties", plan.MaxProperties))
		}
	case "unit":
		if units >= plan.MaxUnits {
			return apperrors.BadRequest(fmt.Sprintf("plan limit reached: max %d units", plan.MaxUnits))
		}
	case "user":
		if users >= plan.MaxUsers {
			return apperrors.BadRequest(fmt.Sprintf("plan limit reached: max %d users", plan.MaxUsers))
		}
	}
	return nil
}
