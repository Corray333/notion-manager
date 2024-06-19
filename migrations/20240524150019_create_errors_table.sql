-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS errors(
    project_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    type TEXT NOT NULL,
    message TEXT NOT NULL,
    page_id TEXT NOT NULL,
    PRIMARY KEY (project_id, page_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE errors;
-- +goose StatementEnd
