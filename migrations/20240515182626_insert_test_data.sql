-- +goose Up
-- +goose StatementBegin
INSERT INTO projects (project_id, name, time_db_id, tasks_db_id, workers_db_id, last_synced) VALUES ('d1f2de35221f4d01a27fcdbda5464fd6', 'Экомобайл', '4ca9a281ae6d49e7b859279809a30401', 'd98dbaea895f4fdebf3d2162d4db54f1', '6eff59b93453498ca6087246c8ae186d', 0);
INSERT INTO ids VALUES ('e754753f491b4fd58913d1fc51ce2f12','d1f2de35221f4d01a27fcdbda5464fd6');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
