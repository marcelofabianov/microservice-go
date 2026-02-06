-- +goose Up
-- +goose StatementBegin

CREATE TYPE user_role AS ENUM ('admin', 'common', 'guest');

CREATE TABLE users (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(254) NOT NULL UNIQUE,
    phone VARCHAR(30) NOT NULL UNIQUE,
    hashed_password VARCHAR(254) NOT NULL,
    role user_role NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    audit_created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    audit_created_by VARCHAR(254),
    audit_updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    audit_updated_by VARCHAR(254),
    audit_archived_at TIMESTAMPTZ,
    audit_deleted_at TIMESTAMPTZ,
    audit_version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_users_email ON users (email)
WHERE
    audit_deleted_at IS NULL;

CREATE INDEX idx_users_phone ON users (phone)
WHERE
    audit_deleted_at IS NULL;

CREATE INDEX idx_users_role ON users (role)
WHERE
    audit_deleted_at IS NULL;

CREATE INDEX idx_users_is_active ON users (is_active)
WHERE
    audit_deleted_at IS NULL;

CREATE INDEX idx_users_deleted_at ON users (audit_deleted_at);

CREATE INDEX idx_users_archived_at ON users (audit_archived_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_users_archived_at;

DROP INDEX IF EXISTS idx_users_deleted_at;

DROP INDEX IF EXISTS idx_users_is_active;

DROP INDEX IF EXISTS idx_users_role;

DROP INDEX IF EXISTS idx_users_phone;

DROP INDEX IF EXISTS idx_users_email;

DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS user_role;

-- +goose StatementEnd
