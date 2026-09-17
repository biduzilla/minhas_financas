CREATE TABLE categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL,
    name        TEXT NOT NULL,
    type        SMALLINT NOT NULL CHECK (type IN (0, 1)),
    goal_id     UUID,
    version     INTEGER NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by  UUID,
    updated_at  TIMESTAMPTZ,
    updated_by  UUID,
    deleted     BOOLEAN NOT NULL DEFAULT false
);

CREATE UNIQUE INDEX idx_categories_unique_active
ON categories(user_id, name, type)
WHERE deleted = false;