package storage

import (
	"github.com/Corray333/notion-manager/internal/project"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	DB *sqlx.DB
}

func NewStorage() *Storage {
	return &Storage{DB: MustInit()}
}

func (s *Storage) NewProject(name string, timeDBID string, tasksDBID string) error {
	_, err := s.DB.Exec("INSERT INTO projects (name, time_db_id, tasks_db_id) VALUES (?, ?, ?)", name, timeDBID, tasksDBID)
	return err
}

func (s *Storage) GetProjects() ([]project.Project, error) {
	var projects []project.Project
	err := s.DB.Select(&projects, "SELECT * FROM projects")
	return projects, err
}

func (s *Storage) GetClientID(internalID string) (string, error) {
	var clientID string
	err := s.DB.Get(&clientID, "SELECT client_id FROM ids WHERE internal_id = ?", internalID)
	return clientID, err
}

func (s *Storage) GetInternalID(clientID string) (string, error) {
	var internalID string
	err := s.DB.Get(&internalID, "SELECT internal_id FROM ids WHERE client_id = ?", clientID)
	return internalID, err
}

func (s *Storage) SetClientID(internalID, clientID string) error {
	_, err := s.DB.Exec("INSERT INTO ids (internal_id, client_id) VALUES (?, ?)", internalID, clientID)
	return err
}

func (s *Storage) SetLastSynced(project project.Project) error {
	_, err := s.DB.Exec("UPDATE projects SET tasks_last_synced = ?, time_last_synced = ? WHERE project_id = ?", project.TasksLastSynced, project.TimeLastSynced, project.ProjectID)
	return err
}
