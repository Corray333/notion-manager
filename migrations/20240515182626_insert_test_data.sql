-- +goose Up
-- +goose StatementBegin
INSERT INTO projects (project_id, name, time_db_id, tasks_db_id, workers_db_id, last_synced) VALUES ('5f332e901a34417ab975d21e098a08a7', 'Радио', '4ca9a281ae6d49e7b859279809a30401', 'd98dbaea895f4fdebf3d2162d4db54f1', '6eff59b93453498ca6087246c8ae186d', 0);
INSERT INTO ids VALUES ('619902a09c6e47bcb7b5d6593a1f7dd7','5f332e901a34417ab975d21e098a08a7');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
