// Package handlers provides the http.HandlerFunc implementations for the server.

package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/Corray333/notion-manager/internal/gsheets"
	"github.com/Corray333/notion-manager/internal/mindmap"
	"github.com/Corray333/notion-manager/internal/notion"
	"github.com/Corray333/notion-manager/internal/project"
)

type Storage interface {
	NewProject(proj *project.Project) error
	GetProjects() ([]project.Project, error)
	SetLastSynced(project *project.Project) error
	GetClientID(internalID string) (string, error)
	GetInternalID(clientID string) (string, error)
	SetClientID(internalID, clientID string) error
	SaveError(err notion.Error) error
	SaveRowsToBeUpdated(notion.Validation)
	GetRowsToBeUpdated() ([]notion.Validation, error)
	GetRowsToBeUpdatedByProject(projectID string) ([]notion.Validation, error)
	RemoveRowToBeUpdated(internalID string) error
}

type NewProjectRequest struct {
	Name              string `json:"name"`                // Project name
	ProjectClientId   string `json:"project_client_id"`   // ID of project in internal dashboard
	ProjectInternalID string `json:"project_internal_id"` // ID of project in client dashboard
	TimeDBID          string `json:"time_db_id"`          // ID of time database in client dashboard
	TasksDBID         string `json:"tasks_db_id"`         // ID of tasks database in client dashboard
	TimeLastSynced    int    `json:"time_last_synced"`    // Time to start searching for updates in time database
	TasksLastSynced   int    `json:"tasks_last_synced"`   // Time to start searching for updates in tasks database
	WorkerDBID        string `json:"worker_db_id"`        // ID of worker database in client dashboard
}

// // NewProject creates a new project
// // @Summary Create a new project
// // @Description Create a new project with the given details
// // @Tags projects
// // @Accept  json
// // @Produce  json
// // @Param   project body NewProjectRequest true "New Project"
// // @Success 201 {string} string "Created"
// // @Failure 400 {string} string "Bad Request"
// // @Failure 500 {string} string "Internal Server Error"
// // @Router /projects [post]
// func NewProject(store Storage) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		var req NewProjectRequest
// 		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 			slog.Error("error decoding request: " + err.Error())
// 			http.Error(w, err.Error(), http.StatusBadRequest)
// 			return
// 		}
// 		proj := project.Project{
// 			Name:        req.Name,
// 			ProjectID:   req.ProjectClientId,
// 			InternalID:  req.ProjectInternalID,
// 			TimeDBID:    req.TimeDBID,
// 			TasksDBID:   req.TasksDBID,
// 			WorkersDBID: req.WorkerDBID,
// 		}
// 		if err := store.NewProject(proj); err != nil {
// 			slog.Error("error creating project: " + err.Error())
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		w.WriteHeader(http.StatusCreated)
// 	}
// }

// UpdateDatabases triggers the update of databases
// @Summary Update databases
// @Description Start the process of updating the databases
// @Tags databases
// @Success 202 {string} string "Accepted"
// @Failure 500 {string} string "Internal Server Error"
// @Router /sync [patch]
func UpdateDatabases(store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: forbid multiple updates at the same time
		if err := notion.StartSync(store); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

// GetToBeUpdated retrieves the rows that need to be updated
// @Summary Get rows to be updated
// @Description Retrieve the rows that need to be updated
// @Tags updates
// @Produce  json
// @Success 200 {array} notion.Validation "OK"
// @Failure 500 {string} string "Internal Server Error"
// @Router /fix [get]
func GetToBeUpdated(store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := store.GetRowsToBeUpdated()
		if err != nil {
			slog.Error("error getting rows to be updated: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(rows); err != nil {
			slog.Error("error encoding response: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return

		}
	}
}

func UpdateGoogleSheets(w http.ResponseWriter, r *http.Request) {
	err := gsheets.UpdateGoogleSheets()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error(err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)

}

func ParseMindmap(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		slog.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		slog.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	projectName, tasks, err := mindmap.ParseMarkdownTasks(string(data))
	if err != nil {
		slog.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := notion.CreateMindmapTasks(projectName, tasks); err != nil {
		slog.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
