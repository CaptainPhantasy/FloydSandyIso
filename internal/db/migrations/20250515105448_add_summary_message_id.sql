-- +goose Up
-- +goose StatementBegin
-- Check if column exists before adding (SQLite compatibility)
-- This migration has already been applied in some instances
-- The column summary_message_id should exist if this migration was already run
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sessions DROP COLUMN summary_message_id;
-- +goose StatementEnd
