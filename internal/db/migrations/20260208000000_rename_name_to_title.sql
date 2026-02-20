-- +goose Up
-- +goose StatementBegin
-- No-op: name -> title migration is handled by ensureColumns in connect.go
-- which safely checks for the name column before migrating.
SELECT 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- No down migration needed - keeping both columns for safety
-- +goose StatementEnd
