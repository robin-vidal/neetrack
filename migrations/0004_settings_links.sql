-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_settings (
    user_id          BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    target_retention DOUBLE PRECISION NOT NULL DEFAULT 0.90
                     CHECK (target_retention BETWEEN 0.70 AND 0.99),
    target_date      DATE,
    block_size       SMALLINT NOT NULL DEFAULT 4 CHECK (block_size BETWEEN 1 AND 6),
    open_max         SMALLINT NOT NULL DEFAULT 1 CHECK (open_max BETWEEN 1 AND 3),
    new_pct          SMALLINT NOT NULL DEFAULT 40 CHECK (new_pct BETWEEN 0 AND 100),
    session_size     SMALLINT NOT NULL DEFAULT 5 CHECK (session_size BETWEEN 1 AND 20)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE problem_links (
    user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_slug  TEXT NOT NULL,
    neetcode_slug TEXT NOT NULL,
    PRIMARY KEY (user_id, problem_slug)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE problem_links;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE user_settings;
-- +goose StatementEnd
