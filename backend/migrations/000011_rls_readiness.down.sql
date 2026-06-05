-- 000011_rls_readiness.down.sql

DROP POLICY IF EXISTS tenant_isolation_properties ON properties;
DROP POLICY IF EXISTS tenant_isolation_units ON units;
DROP POLICY IF EXISTS tenant_isolation_unit_types ON unit_types;
DROP POLICY IF EXISTS tenant_isolation_guests ON guests;
DROP POLICY IF EXISTS tenant_isolation_guest_documents ON guest_documents;
DROP POLICY IF EXISTS tenant_isolation_reservations ON reservations;
DROP POLICY IF EXISTS tenant_isolation_folios ON folios;
DROP POLICY IF EXISTS tenant_isolation_line_items ON line_items;
DROP POLICY IF EXISTS tenant_isolation_invoices ON invoices;
DROP POLICY IF EXISTS tenant_isolation_payments ON payments;
DROP POLICY IF EXISTS tenant_isolation_housekeeping ON housekeeping_tasks;
DROP POLICY IF EXISTS tenant_isolation_laundry_orders ON laundry_orders;
DROP POLICY IF EXISTS tenant_isolation_laundry_items ON laundry_items;
DROP POLICY IF EXISTS tenant_isolation_notification_templates ON notification_templates;
DROP POLICY IF EXISTS tenant_isolation_notification_logs ON notification_logs;
DROP POLICY IF EXISTS tenant_isolation_check_in_details ON check_in_details;

ALTER TABLE properties DISABLE ROW LEVEL SECURITY;
ALTER TABLE units DISABLE ROW LEVEL SECURITY;
ALTER TABLE unit_types DISABLE ROW LEVEL SECURITY;
ALTER TABLE guests DISABLE ROW LEVEL SECURITY;
ALTER TABLE guest_documents DISABLE ROW LEVEL SECURITY;
ALTER TABLE reservations DISABLE ROW LEVEL SECURITY;
ALTER TABLE folios DISABLE ROW LEVEL SECURITY;
ALTER TABLE line_items DISABLE ROW LEVEL SECURITY;
ALTER TABLE invoices DISABLE ROW LEVEL SECURITY;
ALTER TABLE payments DISABLE ROW LEVEL SECURITY;
ALTER TABLE housekeeping_tasks DISABLE ROW LEVEL SECURITY;
ALTER TABLE laundry_orders DISABLE ROW LEVEL SECURITY;
ALTER TABLE laundry_items DISABLE ROW LEVEL SECURITY;
ALTER TABLE notification_templates DISABLE ROW LEVEL SECURITY;
ALTER TABLE notification_logs DISABLE ROW LEVEL SECURITY;
ALTER TABLE check_in_details DISABLE ROW LEVEL SECURITY;

DROP FUNCTION IF EXISTS current_tenant_id();
