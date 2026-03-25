-- name: CreateAccount :one
INSERT INTO accounts (
        user_id,
        currency,
        balance,
        hold_amount,
        status,
        created_at,
        updated_at
    )
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;
-- name: GetAccountByID :one
SELECT *
FROM accounts
WHERE id = $1;
-- name: GetAccountByIDForUpdate :one
SELECT *
FROM accounts
WHERE id = $1 FOR
UPDATE;
-- name: GetAccountsByUserID :many
SELECT *
FROM accounts
WHERE user_id = $1
ORDER BY created_at;
-- name: GetAccountByUserAndCurrency :one
SELECT *
FROM accounts
WHERE user_id = $1
    AND currency = $2;
-- name: UpdateAccount :one
UPDATE accounts
SET balance = $2,
    hold_amount = $3,
    status = $4,
    updated_at = $5
WHERE id = $1
RETURNING *;
-- name: ExistsAccountByUserAndCurrency :one
SELECT EXISTS(
        SELECT 1
        FROM accounts
        WHERE user_id = $1
            AND currency = $2
    );