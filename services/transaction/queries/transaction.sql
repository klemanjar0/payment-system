-- name: CreateTransaction :one
INSERT INTO transactions (
    idempotency_key,
    from_account_id,
    to_account_id,
    amount,
    currency,
    description,
    status
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, idempotency_key, from_account_id, to_account_id, amount, currency, description, status, failure_reason, created_at, updated_at;

-- name: GetTransactionByID :one
SELECT id, idempotency_key, from_account_id, to_account_id, amount, currency, description, status, failure_reason, created_at, updated_at
FROM transactions
WHERE id = $1;

-- name: GetTransactionByIdempotencyKey :one
SELECT id, idempotency_key, from_account_id, to_account_id, amount, currency, description, status, failure_reason, created_at, updated_at
FROM transactions
WHERE idempotency_key = $1;

-- name: GetTransactionsByAccount :many
SELECT id, idempotency_key, from_account_id, to_account_id, amount, currency, description, status, failure_reason, created_at, updated_at
FROM transactions
WHERE from_account_id = $1 OR to_account_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountTransactionsByAccount :one
SELECT COUNT(*)
FROM transactions
WHERE from_account_id = $1 OR to_account_id = $1;

-- name: UpdateTransactionStatus :one
UPDATE transactions
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, idempotency_key, from_account_id, to_account_id, amount, currency, description, status, failure_reason, created_at, updated_at;

-- name: UpdateTransactionStatusWithReason :one
UPDATE transactions
SET status = $2, failure_reason = $3, updated_at = NOW()
WHERE id = $1
RETURNING id, idempotency_key, from_account_id, to_account_id, amount, currency, description, status, failure_reason, created_at, updated_at;
