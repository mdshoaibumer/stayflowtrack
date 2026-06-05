-- 000012_hospitality_operations.up.sql
-- Extended hospitality operations: no-show, maintenance blocking, stay extension,
-- room moves, deposit workflows, and corporate reservations.

-- No-show tracking
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS is_no_show BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS no_show_at TIMESTAMPTZ;
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS no_show_charge DECIMAL(10, 2) DEFAULT 0;

-- Maintenance blocking (room out-of-order for date ranges)
CREATE TABLE IF NOT EXISTS maintenance_blocks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    unit_id UUID NOT NULL REFERENCES units(id) ON DELETE CASCADE,
    reason VARCHAR(500) NOT NULL,
    block_type VARCHAR(30) NOT NULL DEFAULT 'maintenance' CHECK (block_type IN ('maintenance', 'renovation', 'deep_cleaning', 'owner_use', 'other')),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_by UUID REFERENCES users(id),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_maintenance_dates CHECK (end_date >= start_date)
);

CREATE INDEX idx_maintenance_blocks_tenant ON maintenance_blocks(tenant_id);
CREATE INDEX idx_maintenance_blocks_unit_dates ON maintenance_blocks(unit_id, start_date, end_date);
CREATE INDEX idx_maintenance_blocks_property ON maintenance_blocks(property_id, start_date, end_date);

CREATE TRIGGER update_maintenance_blocks_updated_at BEFORE UPDATE ON maintenance_blocks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Stay extension log
CREATE TABLE IF NOT EXISTS stay_extensions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    reservation_id UUID NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
    original_check_out DATE NOT NULL,
    new_check_out DATE NOT NULL,
    additional_nights INT NOT NULL,
    rate_per_night DECIMAL(10, 2) NOT NULL,
    additional_amount DECIMAL(12, 2) NOT NULL,
    reason TEXT,
    extended_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stay_extensions_reservation ON stay_extensions(reservation_id);
CREATE INDEX idx_stay_extensions_tenant ON stay_extensions(tenant_id);

-- Room move log
CREATE TABLE IF NOT EXISTS room_moves (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    reservation_id UUID NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
    from_unit_id UUID NOT NULL REFERENCES units(id),
    to_unit_id UUID NOT NULL REFERENCES units(id),
    reason VARCHAR(500) NOT NULL,
    rate_change DECIMAL(10, 2) DEFAULT 0,
    moved_by UUID REFERENCES users(id),
    moved_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_room_moves_reservation ON room_moves(reservation_id);
CREATE INDEX idx_room_moves_tenant ON room_moves(tenant_id);

-- Corporate accounts
CREATE TABLE IF NOT EXISTS corporate_accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    company_name VARCHAR(255) NOT NULL,
    contact_person VARCHAR(255),
    email VARCHAR(255),
    phone VARCHAR(20),
    gst_number VARCHAR(20),
    billing_address TEXT,
    credit_limit DECIMAL(12, 2) DEFAULT 0,
    payment_terms_days INT DEFAULT 30,
    negotiated_rate_discount DECIMAL(5, 2) DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_corporate_accounts_tenant ON corporate_accounts(tenant_id);
CREATE INDEX idx_corporate_accounts_active ON corporate_accounts(tenant_id) WHERE is_active = true;

CREATE TRIGGER update_corporate_accounts_updated_at BEFORE UPDATE ON corporate_accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Link reservations to corporate accounts
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS corporate_account_id UUID REFERENCES corporate_accounts(id);
CREATE INDEX idx_reservations_corporate ON reservations(corporate_account_id) WHERE corporate_account_id IS NOT NULL;

-- Deposit tracking (explicit table for refund workflow)
CREATE TABLE IF NOT EXISTS deposits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    reservation_id UUID NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
    folio_id UUID REFERENCES folios(id),
    amount DECIMAL(10, 2) NOT NULL,
    method VARCHAR(20) NOT NULL CHECK (method IN ('cash', 'upi', 'card', 'bank_transfer')),
    reference_number VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'held' CHECK (status IN ('held', 'applied', 'refunded', 'forfeited')),
    collected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    collected_by UUID REFERENCES users(id),
    released_at TIMESTAMPTZ,
    released_by UUID REFERENCES users(id),
    refund_amount DECIMAL(10, 2),
    refund_method VARCHAR(20) CHECK (refund_method IN ('cash', 'upi', 'card', 'bank_transfer')),
    refund_reference VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_deposits_tenant ON deposits(tenant_id);
CREATE INDEX idx_deposits_reservation ON deposits(reservation_id);
CREATE INDEX idx_deposits_held ON deposits(tenant_id, status) WHERE status = 'held';

CREATE TRIGGER update_deposits_updated_at BEFORE UPDATE ON deposits
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
