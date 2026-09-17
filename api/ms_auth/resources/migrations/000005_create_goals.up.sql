CREATE TABLE goals (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL,
    name           TEXT NOT NULL,
    target_amount  BIGINT NOT NULL CHECK (target_amount > 0),
    current_amount BIGINT NOT NULL DEFAULT 0 CHECK (current_amount >= 0),
    status         INTEGER NOT NULL DEFAULT 0 CHECK (status BETWEEN 0 AND 3),
    deadline       DATE NOT NULL,
    description    TEXT,
    version        INTEGER NOT NULL DEFAULT 1,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by     UUID,
    updated_at     TIMESTAMPTZ,
    updated_by     UUID,
    deleted        BOOLEAN NOT NULL DEFAULT false
);

CREATE UNIQUE INDEX goal_user_id_name_type_key
    ON goals(user_id, name)
    WHERE deleted = false;

CREATE INDEX idx_goals_user     ON goals(user_id)     WHERE deleted = false;
CREATE INDEX idx_goals_deadline ON goals(deadline)    WHERE deleted = false;
CREATE INDEX idx_goals_status   ON goals(status)      WHERE deleted = false;