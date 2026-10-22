-- name: CreateDefaultSettings :exec
INSERT INTO user_settings (user_id)
VALUES ($1);

-- name: SettingsByUser :one
SELECT target_retention, target_date, block_size, open_max, new_pct, session_size, include_hard
FROM user_settings
WHERE user_id = $1;

-- name: UpdateSettings :exec
UPDATE user_settings
SET target_retention = $2,
    target_date = $3,
    block_size = $4,
    open_max = $5,
    new_pct = $6,
    session_size = $7,
    include_hard = $8
WHERE user_id = $1;
