CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    password_hash BYTEA NOT NULL,
    activated   BOOLEAN NOT NULL DEFAULT false,
    version     INTEGER NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by  UUID,
    updated_at  TIMESTAMPTZ,
    updated_by  UUID,
    deleted     BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted = false;