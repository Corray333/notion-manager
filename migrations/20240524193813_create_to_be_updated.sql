-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS to_be_updated(
    title TEXT NOT NULL,
    type TEXT NOT NULL,
    internal_id TEXT NOT NULL,
    client_id TEXT NOT NULL,
    errors TEXT NOT NULL,
    PRIMARY KEY (internal_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE to_be_updated;
-- +goose StatementEnd
