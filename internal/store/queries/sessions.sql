-- name: CreateSession :exec
INSERT INTO sessions (id, user_id, expires_at)
VALUES ($1, $2, $3);

-- name: SessionWithUserByID :one
SELECT sessions.id, sessions.user_id, sessions.expires_at, users.name AS user_name
FROM sessions
JOIN users ON users.id = sessions.user_id
WHERE sessions.id = $1;

-- name: ExtendSession :exec
UPDATE sessions
SET expires_at = $2
WHERE id = $1;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;
