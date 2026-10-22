-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN is_admin BOOLEAN NOT NULL DEFAULT false;
UPDATE users SET is_admin = true WHERE id = (SELECT id FROM users ORDER BY id LIMIT 1);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN is_admin;
-- +goose StatementEnd
