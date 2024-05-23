-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS projects (
	project_id VARCHAR(80) PRIMARY KEY,
    name VARCHAR(80) ,
	time_db_id VARCHAR(80) ,
	tasks_db_id VARCHAR(80) ,
	workers_db_id VARCHAR(80),
	tasks_last_synced BIGINT,
	time_last_synced BIGINT

);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS projects;
-- +goose StatementEnd
