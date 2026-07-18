-- name: GetOrganization :one
SELECT * FROM "organization" WHERE id = $1;

-- name: GetOrganizationBySlug :one
SELECT * FROM "organization" WHERE slug = $1;

-- name: ListOrganizationsByUserID :many
SELECT o.* FROM "organization" o
JOIN "member" m ON m.organization_id = o.id
WHERE m.user_id = $1 ORDER BY o.name;

-- name: ListOrganizations :many
SELECT * FROM organization ORDER BY name;

-- name: CreateOrganization :one
INSERT INTO "organization" (id, name, slug, logo, metadata) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: UpdateOrganization :one
UPDATE "organization" SET name=$2, slug=$3, logo=COALESCE($4, logo), metadata=$5 WHERE id=$1 RETURNING *;

-- name: UpdateOrganizationLogo :one
UPDATE "organization"
SET logo=$2, logo_key=$3, logo_content_type=$4, logo_updated_at=$5
WHERE id=$1
RETURNING *;

-- name: DeleteOrganization :exec
DELETE FROM "organization" WHERE id = $1;
