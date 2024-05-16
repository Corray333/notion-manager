-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS icons (
    name VARCHAR(255) NOT NULL,
    icon TEXT NOT NULL,
    PRIMARY KEY (name)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS icons;
-- +goose StatementEnd
