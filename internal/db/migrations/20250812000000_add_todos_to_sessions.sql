-- +goose Up
-- +goose StatementBegin
-- Skip if column already exists (database may be ahead of code)
-- ALTER TABLE sessions ADD COLUMN todos TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sessions DROP COLUMN todos;
-- +goose StatementEnd
