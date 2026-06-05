-- Housekeeping Module
CREATE TYPE housekeeping_status AS ENUM ('dirty', 'cleaning', 'inspection', 'ready');
CREATE TYPE housekeeping_priority AS ENUM ('low', 'normal', 'high', 'urgent');

CREATE TABLE housekeeping_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    property_id UUID NOT NULL REFERENCES properties(id),
    unit_id UUID NOT NULL REFERENCES units(id),
    assigned_to UUID REFERENCES users(id),
    status housekeeping_status NOT NULL DEFAULT 'dirty',
    priority housekeeping_priority NOT NULL DEFAULT 'normal',
    task_type VARCHAR(50) NOT NULL DEFAULT 'checkout_clean',
    notes TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    inspected_by UUID REFERENCES users(id),
    inspected_at TIMESTAMPTZ,
    estimated_minutes INT DEFAULT 30,
    actual_minutes INT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_housekeeping_tasks_tenant ON housekeeping_tasks(tenant_id);
CREATE INDEX idx_housekeeping_tasks_property ON housekeeping_tasks(tenant_id, property_id, status);
CREATE INDEX idx_housekeeping_tasks_assigned ON housekeeping_tasks(assigned_to, status);
CREATE INDEX idx_housekeeping_tasks_unit ON housekeeping_tasks(unit_id, status);

CREATE TRIGGER update_housekeeping_tasks_updated_at
    BEFORE UPDATE ON housekeeping_tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
