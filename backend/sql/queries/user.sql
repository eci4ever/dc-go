-- name: GetUser :one
SELECT * FROM "user" WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM "user" WHERE email = $1;

-- name: ListUsers :many
SELECT * FROM "user" ORDER BY name;

-- name: CreateUser :one
INSERT INTO "user" (id, name, email, image, role)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: UpdateUser :one
UPDATE "user" SET name=$2, email=$3, image=$4, role=$5, updated_at=NOW()
WHERE id=$1 RETURNING *;

-- name: DeleteUser :exec
DELETE FROM "user" WHERE id = $1;
