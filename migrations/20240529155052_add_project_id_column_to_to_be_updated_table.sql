-- +goose Up
-- +goose StatementBegin
ALTER TABLE to_be_updated ADD COLUMN project_id TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE to_be_updated DROP COLUMN project_id;
-- +goose StatementEnd
