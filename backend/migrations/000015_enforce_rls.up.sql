-- 000015_enforce_rls.up.sql
-- Enforce Row Level Security by creating a restricted application role.
-- The application connects as stayflow_app (non-owner), which is subject to RLS policies.
-- This closes the defense-in-depth gap where the table owner bypassed RLS.

-- Create restricted application role (non-owner)
-- NOTE: The password MUST be set via environment variable before running this migration.
-- Use: ALTER ROLE stayflow_app PASSWORD 'your-strong-password-here'; after migration.
-- The docker-compose and deploy scripts handle this automatically via STAYFLOW_APP_DB_PASSWORD.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'stayflow_app') THEN
        CREATE ROLE stayflow_app LOGIN IN ROLE stayflow;
        -- Password must be set separately via ALTER ROLE after creation.
        -- This prevents hardcoded credentials in version control.
        RAISE NOTICE 'Created stayflow_app role. Set password with: ALTER ROLE stayflow_app PASSWORD ''<strong-password>'';';
    END IF;
END
$$;

-- Grant necessary permissions to the app role
GRANT USAGE ON SCHEMA public TO stayflow_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO stayflow_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO stayflow_app;

-- Ensure future tables also get granted
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO stayflow_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO stayflow_app;

-- Now FORCE RLS on all tenant-scoped tables so non-owner roles are subject to policies
ALTER TABLE properties FORCE ROW LEVEL SECURITY;
ALTER TABLE units FORCE ROW LEVEL SECURITY;
ALTER TABLE unit_types FORCE ROW LEVEL SECURITY;
ALTER TABLE guests FORCE ROW LEVEL SECURITY;
ALTER TABLE guest_documents FORCE ROW LEVEL SECURITY;
ALTER TABLE reservations FORCE ROW LEVEL SECURITY;
ALTER TABLE folios FORCE ROW LEVEL SECURITY;
ALTER TABLE line_items FORCE ROW LEVEL SECURITY;
ALTER TABLE invoices FORCE ROW LEVEL SECURITY;
ALTER TABLE payments FORCE ROW LEVEL SECURITY;
ALTER TABLE housekeeping_tasks FORCE ROW LEVEL SECURITY;
ALTER TABLE laundry_orders FORCE ROW LEVEL SECURITY;
ALTER TABLE laundry_items FORCE ROW LEVEL SECURITY;
ALTER TABLE notification_templates FORCE ROW LEVEL SECURITY;
ALTER TABLE notification_logs FORCE ROW LEVEL SECURITY;
ALTER TABLE check_in_details FORCE ROW LEVEL SECURITY;

-- The stayflow_app role must set app.current_tenant_id at connection time.
-- The application's database.SetTenantContext() already does this.

COMMENT ON ROLE stayflow_app IS 'Restricted application role subject to RLS policies. Connect the app with this role in production.';
