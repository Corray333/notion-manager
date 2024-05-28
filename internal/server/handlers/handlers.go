// Package handlers provides the http.HandlerFunc implementations for the server.

package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Corray333/notion-manager/internal/notion"
	"github.com/Corray333/notion-manager/internal/project"
)

type Storage interface {
	NewProject(name string, timeDBID string, tasksDBID string, tasks_ls int, time_ls int) error
	GetProjects() ([]project.Project, error)
	SetLastSynced(project project.Project) error
	GetClientID(internalID string) (string, error)
	GetInternalID(clientID string) (string, error)
	SetClientID(internalID, clientID string) error
	SaveErrors(errs []notion.Error) error
	SaveRowsToBeUpdated(notion.Validation)
	GetRowsToBeUpdated() ([]notion.Validation, error)
}

type NewProjectRequest struct {
	Name            string `json:"name"`
	TimeDBID        string `json:"time_db_id"`
	TasksDBID       string `json:"tasks_db_id"`
	TimeLastSynced  int    `json:"time_last_synced"`
	TasksLastSynced int    `json:"tasks_last_synced"`
}

func NewProject(store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req NewProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("error decoding request: " + err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := store.NewProject(req.Name, req.TimeDBID, req.TasksDBID, req.TasksLastSynced, req.TimeLastSynced); err != nil {
			slog.Error("error creating project: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func UpdateDatabases(store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: forbid multiple updates at the same time
		go func() {
			errs := notion.StartSync(store)
			if len(errs) > 0 {
				if err := store.SaveErrors(errs); err != nil {
					slog.Error("error saving errors: " + err.Error())
					return
				}
			}
		}()
		w.Write([]byte(`{"Dashboard sync started."}`))
	}
}

func GetToBeUpdated(store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := store.GetRowsToBeUpdated()
		if err != nil {
			slog.Error("error getting rows to be updated: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(rows)
	}
}
