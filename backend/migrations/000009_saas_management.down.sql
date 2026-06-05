DROP TABLE IF EXISTS onboarding_checklists;
DROP TABLE IF EXISTS billing_events;
DROP TABLE IF EXISTS tenant_subscriptions;
DROP TABLE IF EXISTS subscription_plans;
DROP TYPE IF EXISTS billing_cycle;
DROP TYPE IF EXISTS subscription_status;
ALTER TABLE tenants DROP COLUMN IF EXISTS onboarding_completed;
ALTER TABLE tenants DROP COLUMN IF EXISTS onboarding_step;
