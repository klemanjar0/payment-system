-- name: CreateHold :one
INSERT INTO holds (
        id,
        account_id,
        transaction_id,
        amount,
        description,
        status,
        created_at
    )
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;
-- name: GetHoldByID :one
SELECT *
FROM holds
WHERE id = $1;
-- name: GetHoldByTransactionID :one
SELECT *
FROM holds
WHERE account_id = $1
    AND transaction_id = $2;
-- name: GetHoldByTransactionIDForUpdate :one
SELECT *
FROM holds
WHERE account_id = $1
    AND transaction_id = $2 FOR
UPDATE;
-- name: GetActiveHoldsByAccountID :many
SELECT *
FROM holds
WHERE account_id = $1
    AND status = 'active'
ORDER BY created_at;
-- name: UpdateHold :one
UPDATE holds
SET status = $2,
    released_at = $3
WHERE id = $1
RETURNING *;