-- name: CreateUser :one
INSERT INTO users (name, password_hash)
VALUES ($1, $2)
RETURNING id, name, password_hash, created_at;

-- name: UserByName :one
SELECT id, name, password_hash, created_at
FROM users
WHERE name = $1;
