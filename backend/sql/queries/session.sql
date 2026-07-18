-- name: GetSessionByToken :one
SELECT * FROM "session" WHERE token = $1;

-- name: GetSessionContextByToken :one
SELECT s.*, COALESCE(
  m.role,
  CASE WHEN s.active_organization_id IS NOT NULL AND u.role = 'admin' THEN 'admin' END,
  ''
) AS active_organization_role
FROM "session" s
JOIN "user" u ON u.id = s.user_id
LEFT JOIN "member" m
  ON m.organization_id = s.active_organization_id
 AND m.user_id = s.user_id
WHERE s.token = $1;

-- name: ListSessionsByUserID :many
SELECT * FROM "session" WHERE user_id = $1 AND expires_at > NOW() ORDER BY created_at DESC;

-- name: CreateSession :one
INSERT INTO "session" (id, expires_at, token, ip_address, user_agent, user_id, active_organization_id, active_team_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;

-- name: UpdateSessionActiveOrganization :one
UPDATE "session" s
SET active_organization_id = $2, active_team_id = NULL, updated_at = NOW()
WHERE s.token = $1
  AND s.user_id = $3
  AND s.expires_at > NOW()
  AND EXISTS (
      SELECT 1 FROM "organization" o
      WHERE o.id = $2
        AND (
          EXISTS (
            SELECT 1 FROM "member" m
            WHERE m.organization_id = o.id AND m.user_id = $3
          )
          OR EXISTS (
            SELECT 1 FROM "user" u
            WHERE u.id = $3 AND u.role = 'admin'
          )
        )
  )
RETURNING *;

-- name: ClearActiveOrganizationForMember :exec
UPDATE "session"
SET active_organization_id = NULL, active_team_id = NULL, updated_at = NOW()
WHERE active_organization_id = $1 AND user_id = $2;

-- name: DeleteSession :exec
DELETE FROM "session" WHERE id = $1;

-- name: DeleteSessionByToken :exec
DELETE FROM "session" WHERE token = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM "session" WHERE expires_at < NOW();
