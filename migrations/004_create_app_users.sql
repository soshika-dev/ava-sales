-- +goose Up
CREATE TABLE IF NOT EXISTS app_users (
    id UUID PRIMARY KEY,
    role TEXT NOT NULL CHECK (role IN ('CUSTOMER', 'TECHNICIAN', 'ADMIN')),
    customer_id UUID NULL,
    technician_id UUID NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT app_users_role_identity_check CHECK (
        (role = 'CUSTOMER' AND customer_id IS NOT NULL)
        OR (role = 'TECHNICIAN' AND technician_id IS NOT NULL)
        OR (role = 'ADMIN')
    )
);

CREATE INDEX IF NOT EXISTS idx_app_users_role ON app_users(role);
CREATE INDEX IF NOT EXISTS idx_app_users_customer_id ON app_users(customer_id);
CREATE INDEX IF NOT EXISTS idx_app_users_technician_id ON app_users(technician_id);

-- demo users for local/dev testing
INSERT INTO app_users (id, role, customer_id, technician_id, is_active)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'CUSTOMER', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', NULL, TRUE),
    ('22222222-2222-2222-2222-222222222222', 'TECHNICIAN', NULL, 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', TRUE),
    ('33333333-3333-3333-3333-333333333333', 'ADMIN', NULL, NULL, TRUE)
ON CONFLICT (id) DO NOTHING;

DROP TRIGGER IF EXISTS trg_app_users_updated_at ON app_users;
CREATE TRIGGER trg_app_users_updated_at
BEFORE UPDATE ON app_users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS trg_app_users_updated_at ON app_users;
DELETE FROM app_users WHERE id IN (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    '33333333-3333-3333-3333-333333333333'
);
DROP TABLE IF EXISTS app_users;
