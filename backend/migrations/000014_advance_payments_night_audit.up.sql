-- Add advance payment tracking to reservations
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS advance_amount DECIMAL(12,2) DEFAULT 0;
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS advance_method VARCHAR(50);
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS advance_reference VARCHAR(255);

-- Night audit / day close table
CREATE TABLE IF NOT EXISTS night_audits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    property_id UUID NOT NULL REFERENCES properties(id),
    audit_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'open', -- open, closed
    
    -- Snapshot metrics
    total_units INT NOT NULL DEFAULT 0,
    occupied_units INT NOT NULL DEFAULT 0,
    occupancy_rate DECIMAL(5,2) DEFAULT 0,
    
    -- Collection snapshot
    cash_collected DECIMAL(12,2) DEFAULT 0,
    upi_collected DECIMAL(12,2) DEFAULT 0,
    card_collected DECIMAL(12,2) DEFAULT 0,
    bank_transfer_collected DECIMAL(12,2) DEFAULT 0,
    total_collected DECIMAL(12,2) DEFAULT 0,
    
    -- Operations snapshot
    check_ins INT DEFAULT 0,
    check_outs INT DEFAULT 0,
    walk_ins INT DEFAULT 0,
    no_shows INT DEFAULT 0,
    cancellations INT DEFAULT 0,
    extensions INT DEFAULT 0,
    
    -- Outstanding
    outstanding_amount DECIMAL(12,2) DEFAULT 0,
    outstanding_folios INT DEFAULT 0,
    
    -- Discrepancies
    discrepancies JSONB DEFAULT '[]',
    notes TEXT,
    
    closed_by UUID REFERENCES users(id),
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(tenant_id, property_id, audit_date)
);

-- Index for lookups
CREATE INDEX idx_night_audits_property_date ON night_audits(property_id, audit_date DESC);
