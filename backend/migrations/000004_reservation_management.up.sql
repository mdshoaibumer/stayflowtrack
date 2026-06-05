-- 000004_reservation_management.up.sql
-- Reservations table

CREATE TABLE reservations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    unit_id UUID NOT NULL REFERENCES units(id) ON DELETE RESTRICT,
    guest_id UUID NOT NULL REFERENCES guests(id) ON DELETE RESTRICT,
    booking_source VARCHAR(30) NOT NULL DEFAULT 'walk_in' CHECK (booking_source IN ('walk_in', 'phone', 'whatsapp', 'booking_com', 'airbnb', 'other')),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'checked_in', 'checked_out', 'cancelled')),
    check_in_date DATE NOT NULL,
    check_out_date DATE NOT NULL,
    actual_check_in TIMESTAMPTZ,
    actual_check_out TIMESTAMPTZ,
    num_guests INT NOT NULL DEFAULT 1,
    rate_per_night DECIMAL(10, 2) NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
    notes TEXT,
    cancellation_reason TEXT,
    cancelled_at TIMESTAMPTZ,
    external_booking_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_dates CHECK (check_out_date > check_in_date)
);

CREATE INDEX idx_reservations_tenant_id ON reservations(tenant_id);
CREATE INDEX idx_reservations_property_id ON reservations(property_id);
CREATE INDEX idx_reservations_unit_id ON reservations(unit_id);
CREATE INDEX idx_reservations_guest_id ON reservations(guest_id);
CREATE INDEX idx_reservations_status ON reservations(status);
CREATE INDEX idx_reservations_dates ON reservations(property_id, unit_id, check_in_date, check_out_date);
CREATE INDEX idx_reservations_booking_source ON reservations(booking_source);

-- Prevent overlapping reservations for the same unit (active reservations only)
CREATE UNIQUE INDEX idx_reservations_no_overlap ON reservations (unit_id, check_in_date, check_out_date)
    WHERE status NOT IN ('cancelled', 'checked_out');

-- Function to check for reservation conflicts
CREATE OR REPLACE FUNCTION check_reservation_conflict(
    p_unit_id UUID,
    p_check_in DATE,
    p_check_out DATE,
    p_exclude_id UUID DEFAULT NULL
) RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM reservations
        WHERE unit_id = p_unit_id
          AND status NOT IN ('cancelled', 'checked_out')
          AND check_in_date < p_check_out
          AND check_out_date > p_check_in
          AND (p_exclude_id IS NULL OR id != p_exclude_id)
    );
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_reservations_updated_at BEFORE UPDATE ON reservations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
