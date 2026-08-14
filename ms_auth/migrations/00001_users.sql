-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash BYTEA NOT NULL,
    roles TEXT[] NOT NULL DEFAULT '{ROLE_CLIENT}',
    activated BOOLEAN NOT NULL DEFAULT FALSE,
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    updated_at TIMESTAMPTZ,
    updated_by UUID
);

-- Índice para otimizar as queries de soft-delete (WHERE deleted = false)
CREATE INDEX idx_users_deleted ON users(deleted);

-- +goose Down
DROP TABLE IF EXISTS users;