-- +goose Up
ALTER TABLE sessions ADD COLUMN total_tokens_summarized INTEGER NOT NULL DEFAULT 0;

-- +goose Down
-- Table alteration is hard to reverse in SQLite without recreations, but we can't easily DROP COLUMN in older versions.
-- For v1.5 we just keep it or leave it as is if we don't care about full rollback of this specific field.
