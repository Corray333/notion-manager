-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS employees (
    employee_id UUID PRIMARY KEY,
    username TEXT NOT NULL DEFAULT '',
    icon TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL DEFAULT ''
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS employees;
-- +goose StatementEnd
