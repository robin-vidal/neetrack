-- name: CreateUser :one
INSERT INTO users (name, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: UserByName :one
SELECT *
FROM users
WHERE name = $1;
