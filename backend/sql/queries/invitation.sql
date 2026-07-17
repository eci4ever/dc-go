-- name: GetInvitation :one
SELECT * FROM "invitation" WHERE id = $1;

-- name: GetInvitationForUpdate :one
SELECT * FROM "invitation" WHERE id = $1 FOR UPDATE;

-- name: ListInvitationsByOrganizationID :many
SELECT * FROM "invitation" WHERE organization_id = $1 ORDER BY created_at DESC;

-- name: ListInvitationsByEmail :many
SELECT * FROM "invitation" WHERE email = $1 ORDER BY created_at DESC;

-- name: CreateInvitation :one
INSERT INTO "invitation" (id, organization_id, email, role, expires_at, inviter_id)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: UpdateInvitationStatus :one
UPDATE "invitation" SET status=$2 WHERE id=$1 RETURNING *;

-- name: DeleteInvitation :exec
DELETE FROM "invitation" WHERE id = $1;
