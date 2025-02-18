-- +goose Up
-- +goose StatementBegin
ALTER TABLE employees ADD COLUMN profile TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE employees DROP COLUMN profile;
-- +goose StatementEnd
