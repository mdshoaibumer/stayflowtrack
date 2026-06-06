-- Laundry rate cards: saved item catalog with default pricing
CREATE TABLE IF NOT EXISTS laundry_rate_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    property_id UUID NOT NULL REFERENCES properties(id),
    item_type VARCHAR(50) NOT NULL,
    item_name VARCHAR(100) NOT NULL,
    default_rate NUMERIC(10,2) NOT NULL CHECK (default_rate > 0),
    service_type VARCHAR(20) NOT NULL DEFAULT 'wash',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_laundry_rate_cards_tenant_property ON laundry_rate_cards(tenant_id, property_id);
CREATE INDEX idx_laundry_rate_cards_active ON laundry_rate_cards(tenant_id, property_id) WHERE is_active = true;

-- Add id_document_number to check_in_details if not exists
ALTER TABLE check_in_details ADD COLUMN IF NOT EXISTS id_document_number VARCHAR(50);
