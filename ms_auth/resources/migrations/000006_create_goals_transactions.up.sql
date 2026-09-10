CREATE TABLE goal_transactions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL,
    goal_id        UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    transaction_id UUID NOT NULL,
    amount         NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    version        INTEGER NOT NULL DEFAULT 1,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by     UUID,
    updated_at     TIMESTAMPTZ,
    updated_by     UUID,
    deleted        BOOLEAN NOT NULL DEFAULT false
);

CREATE UNIQUE INDEX goal_transactions_goal_id_transaction_id_key
    ON goal_transactions(goal_id, transaction_id)
    WHERE deleted = false;

CREATE INDEX idx_goal_transactions_user        ON goal_transactions(user_id)        WHERE deleted = false;
CREATE INDEX idx_goal_transactions_goal        ON goal_transactions(goal_id)        WHERE deleted = false;
CREATE INDEX idx_goal_transactions_transaction ON goal_transactions(transaction_id) WHERE deleted = false;