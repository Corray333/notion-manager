package storage

import (
	"fmt"

	"github.com/Corray333/notion-manager/internal/notion"
	"github.com/Corray333/notion-manager/internal/project"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	DB *sqlx.DB
}

func NewStorage() *Storage {
	return &Storage{DB: MustInit()}
}

func (s *Storage) NewProject(name string, timeDBID string, tasksDBID string, tasks_ls int, time_ls int) error {
	_, err := s.DB.Exec("INSERT INTO projects (name, time_db_id, tasks_db_id, tasks_last_synced, time_last_synced) VALUES (?, ?, ?)", name, timeDBID, tasksDBID, tasks_ls, time_ls)
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

func (s *Storage) SaveErrors(errs []notion.Error) error {
	query := squirrel.Insert("errors").Columns("project_id", "type", "message", "page_id")
	for _, err := range errs {
		query = query.Values(err.Unpack())
	}
	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(sql, args...)
	return err
}

func (s *Storage) SaveRowsToBeUpdated(val notion.Validation) {
	query := squirrel.Insert("to_be_updated").Columns("title", "type", "internal_id", "client_id", "errors").Values(val.Title, val.Type, val.InternalID, val.ClientID, val.Errors)
	sql, args, err := query.ToSql()
	if err != nil {
		return
	}
	// TODO: decide, what to do with that
	if _, err := s.DB.Exec(sql, args...); err != nil {
		fmt.Println("Failed to set task to be updated: ", err)
	}
}

func (s *Storage) GetRowsToBeUpdated() ([]notion.Validation, error) {
	var pages []notion.Validation
	var page notion.Validation
	rows, err := s.DB.Queryx("SELECT * FROM to_be_updated")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		err := rows.StructScan(&page)
		if err != nil {
			return nil, err
		}
		pages = append(pages, page)
	}
	return pages, nil
}
