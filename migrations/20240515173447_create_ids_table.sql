-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS ids (
	internal_id VARCHAR(80),
	client_id VARCHAR(80),
	PRIMARY KEY (internal_id, client_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE ids;
-- +goose StatementEnd
