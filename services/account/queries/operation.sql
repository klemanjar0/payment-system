-- name: CreateOperation :one
INSERT INTO operations (
        id,
        account_id,
        type,
        amount,
        balance_after,
        transaction_id,
        description,
        created_at
    )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;
-- name: GetOperationsByAccountID :many
SELECT *
FROM operations
WHERE account_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
-- name: CountOperationsByAccountID :one
SELECT COUNT(*)
FROM operations
WHERE account_id = $1;
-- name: GetOperationByTransactionIDAndType :one
SELECT *
FROM operations
WHERE account_id = $1
    AND transaction_id = $2
    AND type = $3;