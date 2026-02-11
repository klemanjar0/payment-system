-- name: CreateUser :one
INSERT INTO users (
        email,
        phone,
        password_hash,
        first_name,
        last_name,
        status,
        kyc_status,
        created_at,
        updated_at
    )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
returning *;
-- name: GetByID :one
SELECT id,
    email,
    phone,
    password_hash,
    first_name,
    last_name,
    status,
    kyc_status,
    created_at,
    updated_at
FROM users
WHERE id = $1;
-- name: GetByEmail :one
SELECT id,
    email,
    phone,
    password_hash,
    first_name,
    last_name,
    status,
    kyc_status,
    created_at,
    updated_at
FROM users
WHERE email = $1;
-- name: GetByPhone :one
SELECT id,
    email,
    phone,
    password_hash,
    first_name,
    last_name,
    status,
    kyc_status,
    created_at,
    updated_at
FROM users
WHERE phone = $1;
-- name: DeactivateUser :exec
UPDATE users
set user_status = 'deleted'
WHERE ID = $1;
-- name: ExistsByEmail :one
SELECT EXISTS(
        SELECT 1
        FROM users
        WHERE email = $1
    );
-- name: UpdatePassword :exec
UPDATE users
SET password_hash = $2
WHERE ID = $1;
-- name: UpdateUser :one
UPDATE users
SET email = $2,
    phone = $3,
    password_hash = $4,
    first_name = $5,
    last_name = $6,
    status = $7,
    kyc_status = $8,
    updated_at = $9
WHERE id = $1
returning *;