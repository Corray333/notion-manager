// Package handlers provides the http.HandlerFunc implementations for the server.

package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jomei/notionapi"
)

type Storage interface {
	NewProject(name string, timeDBID string, tasksDBID string) error
}

type NewProjectRequest struct {
	Name      string `json:"name"`
	TimeDBID  string `json:"time_db_id"`
	TasksDBID string `json:"tasks_db_id"`
}

func NewProject(store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req NewProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("error decoding request: " + err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := store.NewProject(req.Name, req.TimeDBID, req.TasksDBID); err != nil {
			slog.Error("error creating project: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func UpdateDatabases(store Storage, notion notionapi.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: forbid multiple updates at the same time

	}
}
