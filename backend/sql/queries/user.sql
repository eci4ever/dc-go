-- name: GetUser :one
SELECT * FROM "user" WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM "user" WHERE email = $1;

-- name: ListUsers :many
SELECT * FROM "user" ORDER BY name;

-- name: CreateUser :one
INSERT INTO "user" (id, name, email, image)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateUser :one
UPDATE "user" SET name=$2, email=$3, image=$4, updated_at=NOW()
WHERE id=$1 RETURNING *;

-- name: UpdateUserRole :one
UPDATE "user" SET role=$2, updated_at=NOW()
WHERE id=$1 RETURNING *;

-- name: DeleteUser :exec
DELETE FROM "user" WHERE id = $1;
