-- name: GetTeamMember :one
SELECT * FROM "team_member" WHERE team_id = $1 AND user_id = $2;

-- name: ListTeamMembers :many
SELECT tm.*, u.name, u.email FROM "team_member" tm
JOIN "user" u ON u.id = tm.user_id
WHERE tm.team_id = $1 ORDER BY u.name;

-- name: CreateTeamMember :one
INSERT INTO "team_member" (id, team_id, user_id, created_at) VALUES ($1, $2, $3, NOW()) RETURNING *;

-- name: DeleteTeamMember :exec
DELETE FROM "team_member" WHERE team_id = $1 AND user_id = $2;
