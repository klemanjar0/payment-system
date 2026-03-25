CREATE TYPE account_status AS ENUM ('active', 'blocked', 'closed');
CREATE TYPE hold_status AS ENUM ('active', 'released', 'executed');
CREATE TYPE operation_type AS ENUM ('credit', 'debit', 'hold', 'hold_release');
CREATE TABLE accounts (
    id UUID PRIMARY KEY default gen_random_uuid(),
    user_id UUID NOT NULL,
    currency VARCHAR(3) NOT NULL,
    balance BIGINT NOT NULL DEFAULT 0,
    hold_amount BIGINT NOT NULL DEFAULT 0,
    status account_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT balance_non_negative CHECK (balance >= 0),
    CONSTRAINT hold_non_negative CHECK (hold_amount >= 0),
    CONSTRAINT hold_lte_balance CHECK (hold_amount <= balance),
    UNIQUE(user_id, currency)
);
CREATE TABLE holds (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id),
    transaction_id UUID NOT NULL,
    amount BIGINT NOT NULL,
    description TEXT,
    status hold_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    released_at TIMESTAMPTZ,
    CONSTRAINT amount_positive CHECK (amount > 0),
    UNIQUE(account_id, transaction_id)
);
CREATE TABLE operations (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id),
    type operation_type NOT NULL,
    amount BIGINT NOT NULL,
    balance_after BIGINT NOT NULL,
    transaction_id UUID,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_accounts_user_id ON accounts(user_id);
CREATE INDEX idx_holds_account_id ON holds(account_id);
CREATE INDEX idx_holds_transaction_id ON holds(transaction_id);
CREATE INDEX idx_holds_status ON holds(status)
WHERE status = 'active';
CREATE INDEX idx_operations_account_id ON operations(account_id);
CREATE INDEX idx_operations_created_at ON operations(created_at DESC);