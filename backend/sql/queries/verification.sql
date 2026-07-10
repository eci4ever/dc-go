-- name: GetVerification :one
SELECT * FROM "verification" WHERE identifier = $1 AND value = $2;

-- name: CreateVerification :one
INSERT INTO "verification" (id, identifier, value, expires_at) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: DeleteVerification :exec
DELETE FROM "verification" WHERE id = $1;

-- name: DeleteExpiredVerifications :exec
DELETE FROM "verification" WHERE expires_at < NOW();
