-- name: GetOrganization :one
SELECT * FROM "organization" WHERE id = $1;

-- name: GetOrganizationBySlug :one
SELECT * FROM "organization" WHERE slug = $1;

-- name: ListOrganizationsByUserID :many
SELECT o.* FROM "organization" o
JOIN "member" m ON m.organization_id = o.id
WHERE m.user_id = $1 ORDER BY o.name;

-- name: CreateOrganization :one
INSERT INTO "organization" (id, name, slug, logo, metadata) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: UpdateOrganization :one
UPDATE "organization" SET name=$2, slug=$3, logo=$4, metadata=$5 WHERE id=$1 RETURNING *;

-- name: DeleteOrganization :exec
DELETE FROM "organization" WHERE id = $1;
