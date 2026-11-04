-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    first_name TEXT NOT NULL DEFAULT '',
    last_name TEXT NOT NULL DEFAULT '',
    preferred_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    timezone TEXT NOT NULL DEFAULT 'UTC',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_unq_users_email ON users (LOWER(email));

CREATE TRIGGER set_user_updated_at BEFORE UPDATE ON users
FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS set_user_updated_at ON users;
DROP INDEX IF EXISTS idx_unq_users_email;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
