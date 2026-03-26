CREATE TABLE IF NOT EXISTS orders (
    id          BIGSERIAL PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    number      TEXT NOT NULL UNIQUE,
    status      TEXT NOT NULL DEFAULT 'NEW',
    accrual     DOUBLE PRECISION,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_number ON orders(number);
CREATE INDEX idx_orders_status ON orders(status) WHERE status IN ('NEW', 'PROCESSING');
