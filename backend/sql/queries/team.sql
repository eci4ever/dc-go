-- name: GetTeam :one
SELECT * FROM "team" WHERE id = $1;

-- name: ListTeamsByOrganizationID :many
SELECT * FROM "team" WHERE organization_id = $1 ORDER BY name;

-- name: CreateTeam :one
INSERT INTO "team" (id, name, organization_id) VALUES ($1, $2, $3) RETURNING *;

-- name: UpdateTeam :one
UPDATE "team" SET name=$2, updated_at=NOW() WHERE id=$1 RETURNING *;

-- name: DeleteTeam :exec
DELETE FROM "team" WHERE id = $1;
