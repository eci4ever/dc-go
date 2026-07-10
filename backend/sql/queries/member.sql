-- name: GetMember :one
SELECT * FROM "member" WHERE organization_id = $1 AND user_id = $2;

-- name: ListMembersByOrganizationID :many
SELECT m.*, u.name, u.email, u.image FROM "member" m
JOIN "user" u ON u.id = m.user_id
WHERE m.organization_id = $1 ORDER BY u.name;

-- name: CreateMember :one
INSERT INTO "member" (id, organization_id, user_id, role) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateMemberRole :one
UPDATE "member" SET role=$3 WHERE organization_id=$1 AND user_id=$2 RETURNING *;

-- name: DeleteMember :exec
DELETE FROM "member" WHERE organization_id = $1 AND user_id = $2;
