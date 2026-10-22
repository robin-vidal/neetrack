-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_settings
    ADD COLUMN include_hard BOOLEAN NOT NULL DEFAULT TRUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user_settings
    DROP COLUMN include_hard;
-- +goose StatementEnd
