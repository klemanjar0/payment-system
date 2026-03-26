CREATE TYPE transaction_status AS ENUM ('pending', 'processing', 'completed', 'failed', 'reversed');

CREATE TABLE transactions (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key  VARCHAR(255) NOT NULL UNIQUE,
    from_account_id  UUID NOT NULL,
    to_account_id    UUID NOT NULL,
    amount           BIGINT NOT NULL,
    currency         VARCHAR(10) NOT NULL,
    description      TEXT NOT NULL DEFAULT '',
    status           transaction_status NOT NULL DEFAULT 'pending',
    failure_reason   TEXT NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT amount_positive CHECK (amount > 0)
);

CREATE INDEX idx_transactions_from_account ON transactions(from_account_id);
CREATE INDEX idx_transactions_to_account ON transactions(to_account_id);
CREATE INDEX idx_transactions_status ON transactions(status);
