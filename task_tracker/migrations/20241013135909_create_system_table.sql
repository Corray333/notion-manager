-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS system (
    id SERIAL PRIMARY KEY,
    tasks_db_last_sync BIGINT NOT NULL DEFAULT 0,
    projects_db_last_sync BIGINT NOT NULL DEFAULT 0,
    employee_db_last_sync BIGINT NOT NULL DEFAULT 0,
    times_db_last_sync BIGINT NOT NULL DEFAULT 0
);
INSERT INTO system (tasks_db_last_sync, projects_db_last_sync, employee_db_last_sync) VALUES (0, 0, 0);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS system;
-- +goose StatementEnd
