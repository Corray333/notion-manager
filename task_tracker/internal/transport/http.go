// @title Task Tracker API
// @version 1.0
// @description API for task tracking using notion
// @BasePath /tracker

package transport

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/Corray333/task_tracker/internal/entities"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Transport struct {
	router  *chi.Mux
	service service
}

type service interface {
	GetUsers() ([]entities.Employee, error)
	GetProjects(userID string) ([]entities.Project, error)
	GetTasks(userID, projectID string) ([]entities.Task, error)
	WriteOfTime(time *entities.TimeMsg) error
}

func New(service service) *Transport {
	router := NewRouter()

	return &Transport{
		service: service,
		router:  router,
	}
}

func NewRouter() *chi.Mux {
	router := chi.NewMux()
	router.Use(middleware.Logger)

	// TODO: get allowed origins, headers and methods from cfg
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Set-Cookie", "Refresh", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300, // Максимальное время кеширования предзапроса (в секундах)
	}))

	return router
}

func (s *Transport) Run() {
	slog.Info("Server is starting...")
	panic(http.ListenAndServe("0.0.0.0:"+os.Getenv("SERVER_PORT"), s.router))
}

func (s *Transport) RegisterRoutes() {

	s.router.Group(func(r chi.Router) {
		r.Use(NewAuthMiddleware())
		r.Get("/tracker/employees", s.getEmployees)
		r.Get("/tracker/projects", s.getProjects)
		r.Get("/tracker/tasks", s.getTasks)
		r.Post("/tracker/time", s.writeOfTime)
	})

	s.router.Get("/tracker/swagger/*", httpSwagger.WrapHandler)

}

// GetEmployees godoc
// @Summary Get all employees
// @Description Retrieves a list of employees.
// @Tags employees
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {array} entities.Employee
// @Failure 500 {string} string "Internal Server Error"
// @Router /tracker/employees [get]
func (t *Transport) getEmployees(w http.ResponseWriter, r *http.Request) {
	users, err := t.service.GetUsers()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting users: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding users: %s", err.Error()), http.StatusInternalServerError)
		return
	}
}

// GetProjects godoc
// @Summary Get projects for a specific user
// @Description Retrieves a list of projects for a user by user_id.
// @Tags projects
// @Param user_id query string true "User ID"
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {array} entities.Project
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tracker/projects [get]
func (t *Transport) getProjects(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	projects, err := t.service.GetProjects(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting projects: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(projects); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding projects: %s", err.Error()), http.StatusInternalServerError)
		return
	}

}

// GetTasks godoc
// @Summary Get tasks for a specific project and user
// @Description Retrieves a list of tasks for a user and project by user_id and project_id.
// @Tags tasks
// @Param user_id query string true "User ID"
// @Param project_id query string true "Project ID"
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {array} entities.Task "List of tasks"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tracker/tasks [get]
func (t *Transport) getTasks(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	projectID := r.URL.Query().Get("project_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	if projectID == "" {
		http.Error(w, "project_id is required", http.StatusBadRequest)
		return
	}

	tasks, err := t.service.GetTasks(userID, projectID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting tasks: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding tasks: %s", err.Error()), http.StatusInternalServerError)
		return
	}
}

// WriteOfTime godoc
// @Summary Record the time spent on a task
// @Description Writes the time spent on a task.
// @Tags time
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param time body entities.TimeMsg true "Time data"
// @Success 201 {string} string "Created"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tracker/time [post]
func (t *Transport) writeOfTime(w http.ResponseWriter, r *http.Request) {
	var time entities.TimeMsg
	if err := json.NewDecoder(r.Body).Decode(&time); err != nil {
		http.Error(w, fmt.Sprintf("Error decoding time: %s", err.Error()), http.StatusBadRequest)
		return
	}

	if err := t.service.WriteOfTime(&time); err != nil {
		http.Error(w, fmt.Sprintf("Error writing time: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func NewAuthMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			authToken := r.Header.Get("Authorization")
			authToken = strings.TrimPrefix(authToken, "Bearer ")
			if authToken != os.Getenv("AUTH_TOKEN") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
