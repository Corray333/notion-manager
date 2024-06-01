-- +goose Up
-- +goose StatementBegin
INSERT INTO projects (project_id, name, time_db_id, tasks_db_id, workers_db_id, tasks_last_synced, time_last_synced) VALUES ('925e48a93ff54b4e99594805b5bfbfed', 'Экомобайл', '4ca9a281ae6d49e7b859279809a30401', 'd98dbaea895f4fdebf3d2162d4db54f1', '6eff59b93453498ca6087246c8ae186d', 0, 0);
INSERT INTO ids VALUES ('00039498fe264f6da05b91373de6c0b3','925e48a93ff54b4e99594805b5bfbfed');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
