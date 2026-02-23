-- +goose Up
ALTER TABLE app_users
    ADD COLUMN IF NOT EXISTS email TEXT,
    ADD COLUMN IF NOT EXISTS username TEXT,
    ADD COLUMN IF NOT EXISTS password_hash TEXT;

UPDATE app_users
SET email = CASE id
    WHEN '11111111-1111-1111-1111-111111111111' THEN 'customer@example.com'
    WHEN '22222222-2222-2222-2222-222222222222' THEN 'tech@example.com'
    WHEN '33333333-3333-3333-3333-333333333333' THEN 'admin@example.com'
    ELSE email
END,
username = CASE id
    WHEN '11111111-1111-1111-1111-111111111111' THEN 'customer1'
    WHEN '22222222-2222-2222-2222-222222222222' THEN 'tech1'
    WHEN '33333333-3333-3333-3333-333333333333' THEN 'admin1'
    ELSE username
END,
password_hash = COALESCE(password_hash, '$2y$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi')
WHERE id IN (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    '33333333-3333-3333-3333-333333333333'
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_app_users_email_lower ON app_users (lower(email));
CREATE UNIQUE INDEX IF NOT EXISTS uq_app_users_username_lower ON app_users (lower(username)) WHERE username IS NOT NULL;

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES app_users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

DROP TRIGGER IF EXISTS trg_refresh_tokens_updated_at ON refresh_tokens;
CREATE TRIGGER trg_refresh_tokens_updated_at
BEFORE UPDATE ON refresh_tokens
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS trg_refresh_tokens_updated_at ON refresh_tokens;
DROP TABLE IF EXISTS refresh_tokens;
DROP INDEX IF EXISTS uq_app_users_email_lower;
DROP INDEX IF EXISTS uq_app_users_username_lower;
ALTER TABLE app_users
    DROP COLUMN IF EXISTS password_hash,
    DROP COLUMN IF EXISTS username,
    DROP COLUMN IF EXISTS email;
