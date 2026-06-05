package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Subscription Plan

type PlanTier string

const (
	TierFree         PlanTier = "free"
	TierStarter      PlanTier = "starter"
	TierProfessional PlanTier = "professional"
	TierEnterprise   PlanTier = "enterprise"
)

type SubscriptionPlan struct {
	ID            uuid.UUID       `json:"id"`
	Name          string          `json:"name"`
	Slug          string          `json:"slug"`
	Description   string          `json:"description"`
	Tier          PlanTier        `json:"tier"`
	PriceMonthly  decimal.Decimal `json:"price_monthly"`
	PriceYearly   decimal.Decimal `json:"price_yearly"`
	Currency      string          `json:"currency"`
	MaxProperties int             `json:"max_properties"`
	MaxUnits      int             `json:"max_units"`
	MaxUsers      int             `json:"max_users"`
	Features      []string        `json:"features"`
	IsActive      bool            `json:"is_active"`
	TrialDays     int             `json:"trial_days"`
	SortOrder     int             `json:"sort_order"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// Tenant Subscription

type SubscriptionStatus string

const (
	SubTrialing  SubscriptionStatus = "trialing"
	SubActive    SubscriptionStatus = "active"
	SubPastDue   SubscriptionStatus = "past_due"
	SubCancelled SubscriptionStatus = "cancelled"
	SubExpired   SubscriptionStatus = "expired"
)

type BillingCycle string

const (
	CycleMonthly BillingCycle = "monthly"
	CycleYearly  BillingCycle = "yearly"
)

type TenantSubscription struct {
	ID                     uuid.UUID          `json:"id"`
	TenantID               uuid.UUID          `json:"tenant_id"`
	PlanID                 uuid.UUID          `json:"plan_id"`
	Status                 SubscriptionStatus `json:"status"`
	BillingCycle           BillingCycle       `json:"billing_cycle"`
	CurrentPeriodStart     time.Time          `json:"current_period_start"`
	CurrentPeriodEnd       time.Time          `json:"current_period_end"`
	TrialEndsAt            *time.Time         `json:"trial_ends_at,omitempty"`
	CancelledAt            *time.Time         `json:"cancelled_at,omitempty"`
	RazorpaySubscriptionID string             `json:"razorpay_subscription_id,omitempty"`
	RazorpayCustomerID     string             `json:"razorpay_customer_id,omitempty"`
	CreatedAt              time.Time          `json:"created_at"`
	UpdatedAt              time.Time          `json:"updated_at"`
	// Joined
	PlanName string `json:"plan_name,omitempty"`
	PlanTier string `json:"plan_tier,omitempty"`
}

// Billing Event

type BillingEvent struct {
	ID                uuid.UUID              `json:"id"`
	TenantID          uuid.UUID              `json:"tenant_id"`
	SubscriptionID    *uuid.UUID             `json:"subscription_id,omitempty"`
	EventType         string                 `json:"event_type"`
	Amount            decimal.Decimal        `json:"amount"`
	Currency          string                 `json:"currency"`
	RazorpayPaymentID string                 `json:"razorpay_payment_id,omitempty"`
	RazorpayOrderID   string                 `json:"razorpay_order_id,omitempty"`
	RazorpaySignature string                 `json:"razorpay_signature,omitempty"`
	Status            string                 `json:"status"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
}

// Onboarding

type OnboardingStep struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	PropertyID  *uuid.UUID `json:"property_id,omitempty"`
	StepKey     string     `json:"step_key"`
	StepName    string     `json:"step_name"`
	IsCompleted bool       `json:"is_completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CompletedBy *uuid.UUID `json:"completed_by,omitempty"`
	SortOrder   int        `json:"sort_order"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Admin overview
type TenantOverview struct {
	TenantID           uuid.UUID  `json:"tenant_id"`
	TenantName         string     `json:"tenant_name"`
	TenantEmail        string     `json:"tenant_email"`
	TenantStatus       string     `json:"tenant_status"`
	PlanName           string     `json:"plan_name"`
	PlanTier           string     `json:"plan_tier"`
	SubscriptionStatus string     `json:"subscription_status"`
	PropertyCount      int        `json:"property_count"`
	UserCount          int        `json:"user_count"`
	UnitCount          int        `json:"unit_count"`
	CreatedAt          time.Time  `json:"created_at"`
	LastLoginAt        *time.Time `json:"last_login_at,omitempty"`
}

type SaaSMetrics struct {
	TotalTenants        int             `json:"total_tenants"`
	ActiveTenants       int             `json:"active_tenants"`
	TrialingTenants     int             `json:"trialing_tenants"`
	PaidTenants         int             `json:"paid_tenants"`
	MRR                 decimal.Decimal `json:"mrr"`
	ARR                 decimal.Decimal `json:"arr"`
	ChurnRate           float64         `json:"churn_rate"`
	AvgRevenuePerTenant decimal.Decimal `json:"avg_revenue_per_tenant"`
	TotalProperties     int             `json:"total_properties"`
	TotalUnits          int             `json:"total_units"`
}

// Inputs

type CreateSubscriptionInput struct {
	TenantID     uuid.UUID `json:"tenant_id" validate:"required"`
	PlanSlug     string    `json:"plan_slug" validate:"required"`
	BillingCycle string    `json:"billing_cycle" validate:"required,oneof=monthly yearly"`
}

type ChangePlanInput struct {
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
	PlanSlug string    `json:"new_plan_slug" validate:"required"`
}

type CancelSubscriptionInput struct {
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
	Reason   string    `json:"reason" validate:"omitempty,max=500"`
}

type OnboardPropertyInput struct {
	TenantID   uuid.UUID `json:"tenant_id" validate:"required"`
	PropertyID uuid.UUID `json:"property_id" validate:"required"`
}

type CompleteStepInput struct {
	TenantID   uuid.UUID `json:"tenant_id" validate:"required"`
	PropertyID uuid.UUID `json:"property_id" validate:"required"`
	StepKey    string    `json:"step_key" validate:"required"`
}
