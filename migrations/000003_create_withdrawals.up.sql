CREATE TABLE IF NOT EXISTS withdrawals (
    id            BIGSERIAL PRIMARY KEY,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_number  TEXT NOT NULL,
    sum           DOUBLE PRECISION NOT NULL,
    processed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_withdrawals_user_id ON withdrawals(user_id);
