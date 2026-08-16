-- +goose Up
-- +goose StatementBegin
CREATE TABLE attempts (
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_slug     TEXT NOT NULL,
    pattern          TEXT NOT NULL,
    rating           SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 4),
    duration_seconds INTEGER,
    used_hint        BOOLEAN NOT NULL DEFAULT false,
    saw_solution     BOOLEAN NOT NULL DEFAULT false,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX ON attempts (user_id, created_at DESC);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX ON attempts (user_id, problem_slug);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE pattern_cards (
    user_id        BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pattern        TEXT NOT NULL,
    stability      DOUBLE PRECISION NOT NULL,
    difficulty     DOUBLE PRECISION NOT NULL,
    due_at         TIMESTAMPTZ NOT NULL,
    last_review_at TIMESTAMPTZ NOT NULL,
    reps           INTEGER NOT NULL DEFAULT 0,
    lapses         INTEGER NOT NULL DEFAULT 0,
    state          SMALLINT NOT NULL,
    PRIMARY KEY (user_id, pattern)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX ON pattern_cards (user_id, due_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE pattern_cards;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE attempts;
-- +goose StatementEnd
