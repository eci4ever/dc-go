-- name: GetSessionByToken :one
SELECT * FROM "session" WHERE token = $1;

-- name: ListSessionsByUserID :many
SELECT * FROM "session" WHERE user_id = $1 ORDER BY created_at DESC;

-- name: CreateSession :one
INSERT INTO "session" (id, expires_at, token, ip_address, user_agent, user_id, active_organization_id)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: DeleteSession :exec
DELETE FROM "session" WHERE id = $1;

-- name: DeleteSessionByToken :exec
DELETE FROM "session" WHERE token = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM "session" WHERE expires_at < NOW();
