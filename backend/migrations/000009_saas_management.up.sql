-- SaaS Management: Subscription Plans & Billing

-- Subscription Plans
CREATE TABLE subscription_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    tier VARCHAR(20) NOT NULL CHECK (tier IN ('free', 'starter', 'professional', 'enterprise')),
    price_monthly DECIMAL(10,2) NOT NULL DEFAULT 0,
    price_yearly DECIMAL(10,2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    max_properties INT NOT NULL DEFAULT 1,
    max_units INT NOT NULL DEFAULT 10,
    max_users INT NOT NULL DEFAULT 3,
    features JSONB NOT NULL DEFAULT '[]',
    is_active BOOLEAN NOT NULL DEFAULT true,
    trial_days INT NOT NULL DEFAULT 14,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed default plans
INSERT INTO subscription_plans (name, slug, tier, price_monthly, price_yearly, max_properties, max_units, max_users, features, trial_days, sort_order) VALUES
('Free', 'free', 'free', 0, 0, 1, 5, 2, '["calendar", "reservations", "guest_management"]', 0, 1),
('Starter', 'starter', 'starter', 999, 9990, 1, 25, 5, '["calendar", "reservations", "guest_management", "billing", "housekeeping", "reports_basic"]', 14, 2),
('Professional', 'professional', 'professional', 2499, 24990, 3, 100, 15, '["calendar", "reservations", "guest_management", "billing", "housekeeping", "laundry", "reports_advanced", "whatsapp", "api_access"]', 14, 3),
('Enterprise', 'enterprise', 'enterprise', 4999, 49990, 10, 500, 50, '["calendar", "reservations", "guest_management", "billing", "housekeeping", "laundry", "reports_advanced", "whatsapp", "api_access", "multi_property", "custom_branding", "priority_support"]', 30, 4);

-- Tenant Subscriptions
CREATE TYPE subscription_status AS ENUM ('trialing', 'active', 'past_due', 'cancelled', 'expired');
CREATE TYPE billing_cycle AS ENUM ('monthly', 'yearly');

CREATE TABLE tenant_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES subscription_plans(id),
    status subscription_status NOT NULL DEFAULT 'trialing',
    billing_cycle billing_cycle NOT NULL DEFAULT 'monthly',
    current_period_start TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    current_period_end TIMESTAMPTZ NOT NULL,
    trial_ends_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    razorpay_subscription_id VARCHAR(255),
    razorpay_customer_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id)
);

-- Billing Events (immutable log of all billing events)
CREATE TABLE billing_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    subscription_id UUID REFERENCES tenant_subscriptions(id),
    event_type VARCHAR(50) NOT NULL,
    amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    razorpay_payment_id VARCHAR(255),
    razorpay_order_id VARCHAR(255),
    razorpay_signature VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'success', 'failed', 'refunded')),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Property Onboarding Checklist
CREATE TABLE onboarding_checklists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID REFERENCES properties(id) ON DELETE CASCADE,
    step_key VARCHAR(50) NOT NULL,
    step_name VARCHAR(100) NOT NULL,
    is_completed BOOLEAN NOT NULL DEFAULT false,
    completed_at TIMESTAMPTZ,
    completed_by UUID REFERENCES users(id),
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, property_id, step_key)
);

-- Add subscription fields to tenants
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS onboarding_completed BOOLEAN DEFAULT false;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS onboarding_step VARCHAR(50) DEFAULT 'plan_selection';

CREATE INDEX idx_subscription_plans_active ON subscription_plans(is_active, sort_order);
CREATE INDEX idx_tenant_subscriptions_tenant ON tenant_subscriptions(tenant_id);
CREATE INDEX idx_tenant_subscriptions_status ON tenant_subscriptions(status);
CREATE INDEX idx_billing_events_tenant ON billing_events(tenant_id, created_at DESC);
CREATE INDEX idx_billing_events_razorpay ON billing_events(razorpay_payment_id);
CREATE INDEX idx_onboarding_checklists_tenant ON onboarding_checklists(tenant_id, property_id);

CREATE TRIGGER update_subscription_plans_updated_at
    BEFORE UPDATE ON subscription_plans
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tenant_subscriptions_updated_at
    BEFORE UPDATE ON tenant_subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
