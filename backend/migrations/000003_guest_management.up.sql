-- 000003_guest_management.up.sql
-- Guests and Guest Documents tables

CREATE TABLE guests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(20) NOT NULL,
    address TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100) DEFAULT 'India',
    pincode VARCHAR(10),
    nationality VARCHAR(100) DEFAULT 'Indian',
    date_of_birth DATE,
    aadhaar_number VARCHAR(12),
    passport_number VARCHAR(20),
    notes TEXT,
    total_stays INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_guests_tenant_id ON guests(tenant_id);
CREATE INDEX idx_guests_phone ON guests(phone);
CREATE INDEX idx_guests_email ON guests(email);
CREATE INDEX idx_guests_name ON guests(tenant_id, last_name, first_name);
CREATE INDEX idx_guests_aadhaar ON guests(aadhaar_number) WHERE aadhaar_number IS NOT NULL;
CREATE INDEX idx_guests_passport ON guests(passport_number) WHERE passport_number IS NOT NULL;

CREATE TABLE guest_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    guest_id UUID NOT NULL REFERENCES guests(id) ON DELETE CASCADE,
    document_type VARCHAR(50) NOT NULL CHECK (document_type IN ('aadhaar', 'passport', 'driving_license', 'voter_id', 'other')),
    file_key VARCHAR(500) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_guest_documents_guest_id ON guest_documents(guest_id);
CREATE INDEX idx_guest_documents_tenant_id ON guest_documents(tenant_id);

CREATE TRIGGER update_guests_updated_at BEFORE UPDATE ON guests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
