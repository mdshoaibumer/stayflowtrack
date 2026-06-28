-- 000016_tenant_rls_runtime_context.up.sql
-- Align RLS policies with the runtime tenant transaction context.

CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS UUID AS $$
BEGIN
    RETURN COALESCE(
        NULLIF(current_setting('app.current_tenant', true), '')::UUID,
        NULLIF(current_setting('app.current_tenant_id', true), '')::UUID
    );
EXCEPTION
    WHEN OTHERS THEN RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

CREATE OR REPLACE FUNCTION is_platform_admin() RETURNS BOOLEAN AS $$
BEGIN
    RETURN COALESCE(NULLIF(current_setting('app.is_platform_admin', true), '')::BOOLEAN, false);
EXCEPTION
    WHEN OTHERS THEN RETURN false;
END;
$$ LANGUAGE plpgsql STABLE;

DROP POLICY IF EXISTS tenant_isolation_properties ON properties;
CREATE POLICY tenant_isolation_properties ON properties
    USING (tenant_id = current_tenant_id() OR is_platform_admin());

DROP POLICY IF EXISTS tenant_isolation_units ON units;
CREATE POLICY tenant_isolation_units ON units
    USING (tenant_id = current_tenant_id() OR is_platform_admin());

DROP POLICY IF EXISTS tenant_isolation_unit_types ON unit_types;
CREATE POLICY tenant_isolation_unit_types ON unit_types
    USING (tenant_id = current_tenant_id() OR is_platform_admin());

DROP POLICY IF EXISTS tenant_isolation_guests ON guests;
CREATE POLICY tenant_isolation_guests ON guests
    USING (tenant_id = current_tenant_id() OR is_platform_admin());

DROP POLICY IF EXISTS tenant_isolation_guest_documents ON guest_documents;
CREATE POLICY tenant_isolation_guest_documents ON guest_documents
    USING (tenant_id = current_tenant_id() OR is_platform_admin());

DROP POLICY IF EXISTS tenant_isolation_reservations ON reservations;
CREATE POLICY tenant_isolation_reservations ON reservations
    USING (tenant_id = current_tenant_id() OR is_platform_admin());

DROP POLICY IF EXISTS tenant_isolation_folios ON folios;
CREATE POLICY tenant_isolation_folios ON folios
    USING (tenant_id = current_tenant_id() OR is_platform_admin());

DROP POLICY IF EXISTS tenant_isolation_line_items ON line_items;
CREATE POLICY tenant_isolation_line_items ON line_items
    USING (tenant_id = current_tenant_id() OR is_platform_admin());

DROP POLICY IF EXISTS tenant_isolation_invoices ON invoices;
CREATE POLICY tenant_isolation_invoices ON invoices
    USING (tenant_id = current_tenant_id() OR is_platform_admin());

DROP POLICY IF EXISTS tenant_isolation_payments ON payments;
CREATE POLICY tenant_isolation_payments ON payments
    USING (tenant_id = current_tenant_id() OR is_platform_admin());

DROP POLICY IF EXISTS tenant_isolation_housekeeping ON housekeeping_tasks;
CREATE POLICY tenant_isolation_housekeeping ON housekeeping_tasks
    USING (tenant_id = current_tenant_id() OR is_platform_admin());

DROP POLICY IF EXISTS tenant_isolation_laundry_orders ON laundry_orders;
CREATE POLICY tenant_isolation_laundry_orders ON laundry_orders
    USING (tenant_id = current_tenant_id() OR is_platform_admin());

DROP POLICY IF EXISTS tenant_isolation_laundry_items ON laundry_items;
CREATE POLICY tenant_isolation_laundry_items ON laundry_items
    USING (order_id IN (SELECT id FROM laundry_orders WHERE tenant_id = current_tenant_id()) OR is_platform_admin());

DROP POLICY IF EXISTS tenant_isolation_notification_templates ON notification_templates;
CREATE POLICY tenant_isolation_notification_templates ON notification_templates
    USING (tenant_id = current_tenant_id() OR is_platform_admin());

DROP POLICY IF EXISTS tenant_isolation_notification_logs ON notification_logs;
CREATE POLICY tenant_isolation_notification_logs ON notification_logs
    USING (tenant_id = current_tenant_id() OR is_platform_admin());

DROP POLICY IF EXISTS tenant_isolation_check_in_details ON check_in_details;
CREATE POLICY tenant_isolation_check_in_details ON check_in_details
    USING (tenant_id = current_tenant_id() OR is_platform_admin());

COMMENT ON FUNCTION current_tenant_id() IS 'Returns the tenant UUID set in transaction-local session variables. Used by RLS policies.';
COMMENT ON FUNCTION is_platform_admin() IS 'Returns true only when the app marks a verified platform-admin transaction.';
