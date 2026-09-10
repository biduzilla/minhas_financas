CREATE TABLE transactions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL,
    amount       NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    category_id  UUID NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    version      INTEGER NOT NULL DEFAULT 1,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by   UUID,
    updated_at   TIMESTAMPTZ,
    updated_by   UUID,
    deleted      BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_transactions_user       ON transactions(user_id)       WHERE deleted = false;
CREATE INDEX idx_transactions_category   ON transactions(category_id)   WHERE deleted = false;
CREATE INDEX idx_transactions_created_at ON transactions(created_at)    WHERE deleted = false;