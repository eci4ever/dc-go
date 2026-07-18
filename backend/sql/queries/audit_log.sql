-- name: CreateOrganizationAuditLog :one
INSERT INTO organization_audit_log (
    id, organization_id, actor_user_id, actor_name, actor_email,
    action, target_type, target_id, details
)
SELECT
    sqlc.arg(id), sqlc.arg(organization_id), u.id, u.name, u.email,
    sqlc.arg(action), sqlc.arg(target_type), sqlc.narg(target_id), sqlc.arg(details)
FROM "user" u
WHERE u.id = sqlc.arg(actor_user_id)
RETURNING *;

-- name: ListOrganizationAuditLogs :many
SELECT * FROM organization_audit_log
WHERE organization_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
