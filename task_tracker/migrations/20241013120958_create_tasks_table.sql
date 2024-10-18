-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tasks(
    task_id TEXT PRIMARY KEY,
    title TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT '',
    employee_id TEXT,
    project_id TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tasks;
-- +goose StatementEnd
