-- name: GetAccountByProvider :one
SELECT * FROM "account" WHERE provider_id = $1 AND account_id = $2;

-- name: GetAccountsByUserID :many
SELECT * FROM "account" WHERE user_id = $1;

-- name: CreateAccount :one
INSERT INTO "account" (id, account_id, provider_id, user_id, password, scope)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: UpdateAccountPassword :one
UPDATE "account" SET password=$2, updated_at=NOW() WHERE id=$1 RETURNING *;
