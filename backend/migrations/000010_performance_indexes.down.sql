-- 000010_performance_indexes.down.sql

DROP INDEX IF EXISTS idx_reservations_active_unit_dates;
DROP INDEX IF EXISTS idx_reservations_calendar;
DROP INDEX IF EXISTS idx_payments_tenant_created;
DROP INDEX IF EXISTS idx_folios_open_reservation;
DROP INDEX IF EXISTS idx_guests_name_trgm;
DROP INDEX IF EXISTS idx_guests_phone_trgm;
DROP INDEX IF EXISTS idx_housekeeping_pending;
DROP INDEX IF EXISTS idx_laundry_active;
DROP INDEX IF EXISTS idx_notification_logs_pending;
DROP INDEX IF EXISTS idx_invoices_unpaid;
DROP INDEX IF EXISTS idx_reservations_checkin_date;
DROP INDEX IF EXISTS idx_reservations_checkout_date;
DROP INDEX IF EXISTS idx_tenant_subscriptions_tenant;
DROP INDEX IF EXISTS idx_refresh_tokens_cleanup;
