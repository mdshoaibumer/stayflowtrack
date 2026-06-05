-- Laundry Module
CREATE TYPE laundry_status AS ENUM ('received', 'washing', 'ready', 'delivered');

CREATE TABLE laundry_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    property_id UUID NOT NULL REFERENCES properties(id),
    reservation_id UUID REFERENCES reservations(id),
    guest_id UUID REFERENCES guests(id),
    folio_id UUID REFERENCES folios(id),
    unit_id UUID REFERENCES units(id),
    order_number VARCHAR(50) NOT NULL,
    order_type VARCHAR(20) NOT NULL DEFAULT 'guest' CHECK (order_type IN ('guest', 'house')),
    status laundry_status NOT NULL DEFAULT 'received',
    total_items INT NOT NULL DEFAULT 0,
    total_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    tax_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    grand_total DECIMAL(10,2) NOT NULL DEFAULT 0,
    notes TEXT,
    received_by UUID NOT NULL REFERENCES users(id),
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    washed_at TIMESTAMPTZ,
    ready_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    delivered_by UUID REFERENCES users(id),
    posted_to_folio BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE laundry_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES laundry_orders(id) ON DELETE CASCADE,
    item_type VARCHAR(50) NOT NULL CHECK (item_type IN ('bedsheet', 'towel', 'pillow_cover', 'blanket', 'curtain', 'shirt', 'trouser', 'dress', 'saree', 'other')),
    description VARCHAR(255),
    quantity INT NOT NULL DEFAULT 1,
    unit_price DECIMAL(10,2) NOT NULL DEFAULT 0,
    amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    service_type VARCHAR(20) NOT NULL DEFAULT 'wash' CHECK (service_type IN ('wash', 'dry_clean', 'iron', 'wash_iron')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Sequence for order numbers
CREATE SEQUENCE laundry_order_seq START 1;
CREATE OR REPLACE FUNCTION generate_laundry_order_number(p_tenant_id UUID) RETURNS VARCHAR AS $$
DECLARE
    seq_val BIGINT;
BEGIN
    seq_val := nextval('laundry_order_seq');
    RETURN 'LDR-' || LPAD(seq_val::TEXT, 6, '0');
END;
$$ LANGUAGE plpgsql;

CREATE INDEX idx_laundry_orders_tenant ON laundry_orders(tenant_id);
CREATE INDEX idx_laundry_orders_property ON laundry_orders(tenant_id, property_id, status);
CREATE INDEX idx_laundry_orders_reservation ON laundry_orders(reservation_id);
CREATE INDEX idx_laundry_orders_guest ON laundry_orders(guest_id);
CREATE INDEX idx_laundry_items_order ON laundry_items(order_id);

CREATE TRIGGER update_laundry_orders_updated_at
    BEFORE UPDATE ON laundry_orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
