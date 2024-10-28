-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS projects (
    project_id TEXT PRIMARY KEY,
    icon TEXT NOT NULL DEFAULT '',
    icon_type VARCHAR(16) DEFAULT '',
    name TEXT NOT NULL,
    status VARCHAR(64) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS projects;
-- +goose StatementEnd
