-- name: UpsertProblemLink :exec
INSERT INTO problem_links (user_id, problem_slug, neetcode_slug)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, problem_slug) DO UPDATE SET
    neetcode_slug = EXCLUDED.neetcode_slug;

-- name: ProblemLinksByUser :many
SELECT problem_slug, neetcode_slug
FROM problem_links
WHERE user_id = $1;
