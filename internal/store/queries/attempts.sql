-- name: CreateAttempt :exec
INSERT INTO attempts (user_id, problem_slug, pattern, rating, duration_seconds, used_hint, saw_solution)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: LastAttemptPerProblem :many
SELECT problem_slug, MAX(created_at)::timestamptz AS last_attempt_at
FROM attempts
WHERE user_id = $1
GROUP BY problem_slug;

-- name: AttemptsByUser :many
SELECT pattern, rating, created_at
FROM attempts
WHERE user_id = $1
ORDER BY created_at;

-- name: DoneCountByPattern :many
SELECT pattern, COUNT(DISTINCT problem_slug) AS done
FROM attempts
WHERE user_id = $1 AND rating >= 2
GROUP BY pattern;
