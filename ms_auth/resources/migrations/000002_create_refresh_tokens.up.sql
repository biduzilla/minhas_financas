CREATE TABLE refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash TEXT NOT NULL UNIQUE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    family     UUID NOT NULL,
    revoked    BOOLEAN NOT NULL DEFAULT false,
    version    INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_at TIMESTAMPTZ,
    updated_by UUID,
    deleted    BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_refresh_tokens_family ON refresh_tokens(family) WHERE deleted = false;
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id) WHERE deleted = false;