CREATE TABLE users (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login          TEXT NOT NULL UNIQUE,
    password_hash  TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE orders (
    id          BIGSERIAL PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    number      TEXT NOT NULL UNIQUE,
    status      TEXT NOT NULL DEFAULT 'NEW',
    accrual     DOUBLE PRECISION,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE withdrawals (
    id            BIGSERIAL PRIMARY KEY,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_number  TEXT NOT NULL,
    sum           DOUBLE PRECISION NOT NULL,
    processed_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
