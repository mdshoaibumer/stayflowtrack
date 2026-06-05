-- 000005_billing_engine.up.sql
-- Folios, Line Items, Invoices, and Payments

-- Folios: Represents a tab/bill for a reservation
CREATE TABLE folios (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    reservation_id UUID NOT NULL REFERENCES reservations(id) ON DELETE RESTRICT,
    guest_id UUID NOT NULL REFERENCES guests(id) ON DELETE RESTRICT,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    folio_number VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'closed', 'void')),
    subtotal DECIMAL(12, 2) NOT NULL DEFAULT 0,
    tax_total DECIMAL(12, 2) NOT NULL DEFAULT 0,
    total_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    paid_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    balance DECIMAL(12, 2) NOT NULL DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ,
    UNIQUE(tenant_id, folio_number)
);

CREATE INDEX idx_folios_tenant_id ON folios(tenant_id);
CREATE INDEX idx_folios_reservation_id ON folios(reservation_id);
CREATE INDEX idx_folios_guest_id ON folios(guest_id);
CREATE INDEX idx_folios_status ON folios(status);

-- Line Items: Individual charges on a folio
CREATE TABLE line_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    folio_id UUID NOT NULL REFERENCES folios(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL CHECK (category IN ('room_charge', 'deposit', 'food_beverage', 'laundry', 'minibar', 'parking', 'spa', 'damage', 'late_checkout', 'extra_bed', 'other')),
    description VARCHAR(500) NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    unit_price DECIMAL(10, 2) NOT NULL,
    amount DECIMAL(12, 2) NOT NULL,
    tax_rate DECIMAL(5, 2) NOT NULL DEFAULT 0 CHECK (tax_rate IN (0, 5, 12, 18, 28)),
    tax_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    total DECIMAL(12, 2) NOT NULL,
    date DATE NOT NULL DEFAULT CURRENT_DATE,
    is_void BOOLEAN NOT NULL DEFAULT false,
    void_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);

CREATE INDEX idx_line_items_folio_id ON line_items(folio_id);
CREATE INDEX idx_line_items_tenant_id ON line_items(tenant_id);
CREATE INDEX idx_line_items_category ON line_items(category);

-- Invoices: Finalized bills (immutable)
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    folio_id UUID NOT NULL REFERENCES folios(id) ON DELETE RESTRICT,
    reservation_id UUID NOT NULL REFERENCES reservations(id) ON DELETE RESTRICT,
    guest_id UUID NOT NULL REFERENCES guests(id) ON DELETE RESTRICT,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    invoice_number VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'issued' CHECK (status IN ('issued', 'paid', 'partially_paid', 'void', 'refunded')),
    guest_name VARCHAR(255) NOT NULL,
    guest_email VARCHAR(255),
    guest_phone VARCHAR(20),
    guest_address TEXT,
    property_name VARCHAR(255) NOT NULL,
    property_address TEXT NOT NULL,
    property_gst_number VARCHAR(20),
    subtotal DECIMAL(12, 2) NOT NULL,
    cgst_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    sgst_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    igst_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    total_tax DECIMAL(12, 2) NOT NULL DEFAULT 0,
    total_amount DECIMAL(12, 2) NOT NULL,
    paid_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    balance_due DECIMAL(12, 2) NOT NULL,
    check_in_date DATE NOT NULL,
    check_out_date DATE NOT NULL,
    num_nights INT NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    due_date DATE,
    notes TEXT,
    pdf_key VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, invoice_number)
);

CREATE INDEX idx_invoices_tenant_id ON invoices(tenant_id);
CREATE INDEX idx_invoices_folio_id ON invoices(folio_id);
CREATE INDEX idx_invoices_reservation_id ON invoices(reservation_id);
CREATE INDEX idx_invoices_guest_id ON invoices(guest_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_issued_at ON invoices(issued_at);

-- Invoice Line Items (snapshot at time of invoicing - immutable)
CREATE TABLE invoice_line_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL,
    description VARCHAR(500) NOT NULL,
    quantity INT NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    amount DECIMAL(12, 2) NOT NULL,
    tax_rate DECIMAL(5, 2) NOT NULL,
    tax_amount DECIMAL(12, 2) NOT NULL,
    total DECIMAL(12, 2) NOT NULL,
    date DATE NOT NULL
);

CREATE INDEX idx_invoice_line_items_invoice_id ON invoice_line_items(invoice_id);

-- Payments
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    folio_id UUID NOT NULL REFERENCES folios(id) ON DELETE RESTRICT,
    invoice_id UUID REFERENCES invoices(id) ON DELETE SET NULL,
    payment_type VARCHAR(20) NOT NULL CHECK (payment_type IN ('payment', 'refund', 'deposit', 'deposit_release')),
    payment_method VARCHAR(20) NOT NULL CHECK (payment_method IN ('cash', 'upi', 'card', 'bank_transfer')),
    amount DECIMAL(12, 2) NOT NULL,
    reference_number VARCHAR(255),
    notes TEXT,
    received_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_tenant_id ON payments(tenant_id);
CREATE INDEX idx_payments_folio_id ON payments(folio_id);
CREATE INDEX idx_payments_invoice_id ON payments(invoice_id);
CREATE INDEX idx_payments_payment_type ON payments(payment_type);
CREATE INDEX idx_payments_created_at ON payments(created_at);

-- Check-in details table
CREATE TABLE check_in_details (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    reservation_id UUID NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
    assigned_unit_id UUID NOT NULL REFERENCES units(id),
    deposit_amount DECIMAL(10, 2) DEFAULT 0,
    deposit_method VARCHAR(20) CHECK (deposit_method IN ('cash', 'upi', 'card', 'bank_transfer')),
    deposit_reference VARCHAR(255),
    id_document_type VARCHAR(50),
    id_document_key VARCHAR(500),
    guest_signature_key VARCHAR(500),
    checked_in_by UUID REFERENCES users(id),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(reservation_id)
);

CREATE INDEX idx_check_in_details_reservation ON check_in_details(reservation_id);
CREATE INDEX idx_check_in_details_tenant ON check_in_details(tenant_id);

-- Sequence for folio numbers (per tenant)
CREATE SEQUENCE folio_number_seq START 1;
CREATE SEQUENCE invoice_number_seq START 1;

-- Function to generate folio number
CREATE OR REPLACE FUNCTION generate_folio_number(p_tenant_id UUID)
RETURNS VARCHAR AS $$
DECLARE
    v_number BIGINT;
BEGIN
    v_number := nextval('folio_number_seq');
    RETURN 'FOL-' || LPAD(v_number::TEXT, 8, '0');
END;
$$ LANGUAGE plpgsql;

-- Function to generate invoice number
CREATE OR REPLACE FUNCTION generate_invoice_number(p_tenant_id UUID)
RETURNS VARCHAR AS $$
DECLARE
    v_number BIGINT;
BEGIN
    v_number := nextval('invoice_number_seq');
    RETURN 'INV-' || TO_CHAR(NOW(), 'YYYYMM') || '-' || LPAD(v_number::TEXT, 6, '0');
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_folios_updated_at BEFORE UPDATE ON folios
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
