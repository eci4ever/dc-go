-- name: GetOrganization :one
SELECT * FROM "organization" WHERE id = $1;

-- name: GetOrganizationBySlug :one
SELECT * FROM "organization" WHERE slug = $1;

-- name: LockOrganization :one
SELECT id FROM organization WHERE id = $1 FOR UPDATE;

-- name: ListOrganizationsByUserID :many
SELECT o.* FROM "organization" o
JOIN "member" m ON m.organization_id = o.id
WHERE m.user_id = $1 ORDER BY o.name;

-- name: ListOwnedOrganizationsByUserID :many
SELECT o.* FROM "organization" o
JOIN "member" m ON m.organization_id = o.id
WHERE m.user_id = $1 AND m.role = 'owner'
ORDER BY o.name;

-- name: ListOrganizations :many
SELECT * FROM organization ORDER BY name;

-- name: ListOrganizationsWithOwner :many
SELECT
    o.*,
    owner_user.id AS owner_id,
    owner_user.name AS owner_name,
    owner_user.email AS owner_email,
    owner_user.image AS owner_image
FROM organization o
LEFT JOIN member owner_member ON owner_member.id = (
    SELECT m.id
    FROM member m
    WHERE m.organization_id = o.id AND m.role = 'owner'
    ORDER BY m.created_at, m.id
    LIMIT 1
)
LEFT JOIN "user" owner_user ON owner_user.id = owner_member.user_id
ORDER BY o.name;

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

-- name: DemoteOrganizationOwners :exec
UPDATE member
SET role = 'admin'
WHERE organization_id = $1 AND role = 'owner' AND user_id <> $2;

-- name: UpsertOrganizationOwner :one
INSERT INTO member (id, organization_id, user_id, role)
VALUES ($1, $2, $3, 'owner')
ON CONFLICT (organization_id, user_id)
DO UPDATE SET role = 'owner'
RETURNING *;

-- name: GetOrganizationOwner :one
SELECT u.id, u.name, u.email, u.image
FROM member m
JOIN "user" u ON u.id = m.user_id
WHERE m.organization_id = $1 AND m.role = 'owner'
ORDER BY m.created_at, m.id
LIMIT 1;
