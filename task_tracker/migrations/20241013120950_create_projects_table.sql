-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS projects (
    project_id UUID PRIMARY KEY,
    icon TEXT NOT NULL DEFAULT '',
    icon_type VARCHAR(16) DEFAULT '',
    name TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS projects;
-- +goose StatementEnd
