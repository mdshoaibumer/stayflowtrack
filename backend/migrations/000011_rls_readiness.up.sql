-- 000011_rls_readiness.up.sql
-- Row Level Security readiness: enables RLS on tenant-scoped tables
-- as a defense-in-depth layer alongside application-level tenant filtering.
--
-- RLS policies use the session variable app.current_tenant_id which must be
-- set by the application at connection time.

-- Helper function to get current tenant
CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS UUID AS $$
BEGIN
    RETURN NULLIF(current_setting('app.current_tenant_id', true), '')::UUID;
EXCEPTION
    WHEN OTHERS THEN RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

-- Enable RLS on all tenant-scoped tables
ALTER TABLE properties ENABLE ROW LEVEL SECURITY;
ALTER TABLE units ENABLE ROW LEVEL SECURITY;
ALTER TABLE unit_types ENABLE ROW LEVEL SECURITY;
ALTER TABLE guests ENABLE ROW LEVEL SECURITY;
ALTER TABLE guest_documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE reservations ENABLE ROW LEVEL SECURITY;
ALTER TABLE folios ENABLE ROW LEVEL SECURITY;
ALTER TABLE line_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;
ALTER TABLE payments ENABLE ROW LEVEL SECURITY;
ALTER TABLE housekeeping_tasks ENABLE ROW LEVEL SECURITY;
ALTER TABLE laundry_orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE laundry_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE notification_templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE notification_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE check_in_details ENABLE ROW LEVEL SECURITY;

-- Create policies (allow all operations when tenant matches)
-- Note: FORCE ROW LEVEL SECURITY is NOT set, so table owners (the app user) bypass RLS.
-- This allows the application to work normally while RLS acts as a safety net.

CREATE POLICY tenant_isolation_properties ON properties
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_units ON units
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_unit_types ON unit_types
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_guests ON guests
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_guest_documents ON guest_documents
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_reservations ON reservations
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_folios ON folios
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_line_items ON line_items
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_invoices ON invoices
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_payments ON payments
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_housekeeping ON housekeeping_tasks
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_laundry_orders ON laundry_orders
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_laundry_items ON laundry_items
    USING (order_id IN (SELECT id FROM laundry_orders WHERE tenant_id = current_tenant_id()));

CREATE POLICY tenant_isolation_notification_templates ON notification_templates
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_notification_logs ON notification_logs
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_check_in_details ON check_in_details
    USING (tenant_id = current_tenant_id());

-- Create a restricted application role for future enforcement
-- When ready to enforce RLS, create a non-owner role and connect the app with it:
--   CREATE ROLE stayflow_app LOGIN PASSWORD '...' IN ROLE stayflow;
--   GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO stayflow_app;
--   ALTER TABLE properties FORCE ROW LEVEL SECURITY;
-- The app user will then be subject to RLS policies.

COMMENT ON FUNCTION current_tenant_id() IS 'Returns the tenant UUID set in the session variable app.current_tenant_id. Used by RLS policies.';
