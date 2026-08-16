-- name: PatternCardsByUser :many
SELECT pattern, stability, difficulty, due_at, last_review_at, reps, lapses, state
FROM pattern_cards
WHERE user_id = $1;

-- name: PatternCardByUserAndPattern :one
SELECT pattern, stability, difficulty, due_at, last_review_at, reps, lapses, state
FROM pattern_cards
WHERE user_id = $1 AND pattern = $2;

-- name: UpsertPatternCard :exec
INSERT INTO pattern_cards (user_id, pattern, stability, difficulty, due_at, last_review_at, reps, lapses, state)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (user_id, pattern) DO UPDATE SET
    stability = EXCLUDED.stability,
    difficulty = EXCLUDED.difficulty,
    due_at = EXCLUDED.due_at,
    last_review_at = EXCLUDED.last_review_at,
    reps = EXCLUDED.reps,
    lapses = EXCLUDED.lapses,
    state = EXCLUDED.state;

-- name: DeletePatternCardsByUser :exec
DELETE FROM pattern_cards
WHERE user_id = $1;
