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

-- name: ActivityCounts :many
SELECT (created_at AT TIME ZONE 'UTC')::date AS day, COUNT(*) AS count
FROM attempts
WHERE user_id = $1 AND created_at >= $2
GROUP BY day
ORDER BY day;

-- name: ProblemAttemptStats :many
SELECT DISTINCT ON (a.problem_slug)
    a.problem_slug,
    a.rating AS last_rating,
    (
        SELECT COUNT(*)
        FROM attempts solved
        WHERE solved.user_id = a.user_id
          AND solved.problem_slug = a.problem_slug
          AND solved.rating >= 3
    )::bigint AS solves
FROM attempts a
WHERE a.user_id = $1
ORDER BY a.problem_slug, a.created_at DESC;

-- name: AttemptLogByUser :many
SELECT problem_slug, pattern, rating, duration_seconds, used_hint, saw_solution, created_at
FROM attempts
WHERE user_id = $1
ORDER BY created_at;

-- name: CreateAttemptAt :exec
INSERT INTO attempts (user_id, problem_slug, pattern, rating, duration_seconds, used_hint, saw_solution, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
