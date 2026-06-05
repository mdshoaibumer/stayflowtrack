-- 000010_performance_indexes.up.sql
-- Performance indexes based on production audit findings

-- Reservation conflict detection (hot path on every create/update)
CREATE INDEX IF NOT EXISTS idx_reservations_active_unit_dates ON reservations(unit_id, check_in_date, check_out_date)
    WHERE status NOT IN ('cancelled', 'checked_out');

-- Calendar view (fetches reservations for property in date range)
CREATE INDEX IF NOT EXISTS idx_reservations_calendar ON reservations(property_id, check_in_date, check_out_date)
    WHERE status NOT IN ('cancelled');

-- Dashboard revenue queries (payments by tenant, date range)
CREATE INDEX IF NOT EXISTS idx_payments_tenant_created ON payments(tenant_id, created_at)
    INCLUDE (amount, payment_type);

-- Open folios lookup (common billing query)
CREATE INDEX IF NOT EXISTS idx_folios_open_reservation ON folios(reservation_id)
    WHERE status = 'open';

-- Guest search (trigram index for ILIKE pattern matching)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_guests_name_trgm ON guests
    USING gin ((first_name || ' ' || last_name) gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_guests_phone_trgm ON guests
    USING gin (phone gin_trgm_ops);

-- Housekeeping: incomplete tasks for dashboard
CREATE INDEX IF NOT EXISTS idx_housekeeping_pending ON housekeeping_tasks(property_id, tenant_id)
    WHERE status NOT IN ('ready');

-- Laundry: active orders
CREATE INDEX IF NOT EXISTS idx_laundry_active ON laundry_orders(property_id, tenant_id)
    WHERE status NOT IN ('delivered');

-- Notifications: pending delivery
CREATE INDEX IF NOT EXISTS idx_notification_logs_pending ON notification_logs(tenant_id, status)
    WHERE status IN ('pending', 'sent');

-- Invoices: unpaid for reporting
CREATE INDEX IF NOT EXISTS idx_invoices_unpaid ON invoices(tenant_id, status)
    WHERE status NOT IN ('paid', 'void');

-- Reservations: today's arrivals/departures (dashboard)
CREATE INDEX IF NOT EXISTS idx_reservations_checkin_date ON reservations(tenant_id, check_in_date)
    WHERE status = 'confirmed';

CREATE INDEX IF NOT EXISTS idx_reservations_checkout_date ON reservations(tenant_id, check_out_date)
    WHERE status = 'checked_in';

-- Subscription lookups
CREATE INDEX IF NOT EXISTS idx_tenant_subscriptions_tenant ON tenant_subscriptions(tenant_id)
    WHERE status IN ('active', 'trialing');

-- Refresh token cleanup (expired tokens)
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_cleanup ON refresh_tokens(expires_at)
    WHERE revoked = false;
