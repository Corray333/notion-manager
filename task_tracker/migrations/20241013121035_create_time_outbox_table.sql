-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS time_outbox (
    time_id BIGINT NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1),
    task_id UUID NOT NULL,
    duration BIGINT NOT NULL DEFAULT 0,
    employee_id UUID NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    CONSTRAINT time_pkey PRIMARY KEY (time_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS time_outbox;
-- +goose StatementEnd
