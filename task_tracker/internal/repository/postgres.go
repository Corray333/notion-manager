package repository

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/Corray333/task_tracker/internal/entities"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Storage struct {
	DB *sqlx.DB
}

func New() *Storage {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB_NAME"))
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	return &Storage{
		DB: db,
	}
}

func (s *Storage) GetEmployees() (employees []entities.Employee, err error) {
	if err := s.DB.Select(&employees, "SELECT * FROM employees"); err != nil {
		slog.Error("error getting employees: " + err.Error())
		return nil, err
	}

	return employees, nil
}

func (s *Storage) GetProjects(userID string) (projects []entities.Project, err error) {
	if err := s.DB.Select(&projects, "SELECT * FROM projects"); err != nil {
		slog.Error("error getting projects: " + err.Error())
		return nil, err
	}

	return projects, nil
}

func (s *Storage) GetTasks(userID string, projectID string) (tasks []entities.Task, err error) {
	statuses := []string{"Формируется", "Можно делать", "На паузе", "Ожидание", "В работе", "Надо обсудить", "Код-ревью", "Внутренняя проверка"}
	query := `
        SELECT * FROM tasks 
        WHERE project_id = $1 
        AND employee_id = $2 
        AND status = ANY($3)
    `
	args := []interface{}{projectID, userID, pq.Array(statuses)}

	if err := s.DB.Select(&tasks, query, args...); err != nil {
		slog.Error("error getting tasks: " + err.Error())
		return nil, err
	}

	return tasks, nil
}

func (s *Storage) SetEmployees(employees []entities.Employee) error {
	tx, err := s.DB.Beginx()
	if err != nil {
		slog.Error("error starting transaction: " + err.Error())
		return err
	}
	defer tx.Rollback()

	for _, employee := range employees {
		_, err := tx.Exec("INSERT INTO employees (employee_id, username, email, icon) VALUES ($1, $2, $3, $4) ON CONFLICT (employee_id) DO UPDATE SET username = $2, email = $3, icon = $4", employee.ID, employee.Username, employee.Email, employee.Icon)
		if err != nil {
			slog.Error("error setting employees: " + err.Error())
			return err
		}
	}

	return tx.Commit()
}

// SetTasks inserts tasks into the postgres database or updates them if they already exist with this uuid
func (s *Storage) SetTasks(tasks []entities.Task) error {
	tx, err := s.DB.Beginx()
	if err != nil {
		slog.Error("error starting transaction: " + err.Error())
		return err
	}
	defer tx.Rollback()

	for _, task := range tasks {
		_, err := tx.Exec("INSERT INTO tasks (task_id, project_id, employee_id, title, status) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (task_id) DO UPDATE SET title = $4, status = $5", task.ID, task.ProjectID, task.EmployeeID, task.Title, task.Status)
		if err != nil {
			slog.Error("error setting tasks: " + err.Error())
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) SetProjects(projects []entities.Project) error {
	tx, err := s.DB.Beginx()
	if err != nil {
		slog.Error("error starting transaction: " + err.Error())
		return err
	}
	defer tx.Rollback()

	for _, project := range projects {
		_, err := tx.Exec("INSERT INTO projects (project_id, name, icon, icon_type) VALUES ($1, $2, $3, $4) ON CONFLICT (project_id) DO UPDATE SET name = $2, icon = $3, icon_type = $4", project.ID, project.Name, project.Icon, project.IconType)
		if err != nil {
			slog.Error("error setting projects: " + err.Error())
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) SaveTimeWriteOf(time *entities.TimeMsg) error {
	_, err := s.DB.Exec("INSERT INTO time_outbox (task_id, employee_id, duration, description) VALUES ($1, $2, $3, $4)", time.TaskID, time.EmployeeID, time.Duration, time.Description)
	if err != nil {
		slog.Error("error saving time write of: " + err.Error())
		return err
	}

	return nil
}

func (s *Storage) GetTimes() (times []entities.TimeMsg, err error) {
	if err = s.DB.Select(&times, "SELECT * FROM time_outbox"); err != nil {
		slog.Error("error getting time outbox messages: " + err.Error())
		return nil, err
	}

	return times, nil
}

func (s *Storage) MarkTimeAsSent(timeID int64) error {
	if _, err := s.DB.Exec("DELETE FROM time_outbox WHERE time_id = $1", timeID); err != nil {
		slog.Error("error marking time as sent: " + err.Error())
		return err
	}

	return nil
}

func (s *Storage) GetSystemInfo() (*entities.System, error) {
	system := entities.System{}
	if err := s.DB.Get(&system, "SELECT * FROM system"); err != nil {
		slog.Error("error getting system info: " + err.Error())
		return nil, err
	}

	return &system, nil
}

func (s *Storage) SetSystemInfo(system *entities.System) error {
	tx, err := s.DB.Beginx()
	if err != nil {
		slog.Error("error starting transaction: " + err.Error())
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE system SET projects_db_last_sync = $1, tasks_db_last_sync = $2, employee_db_last_sync = $3", system.ProjectsDBLastSynced, system.TasksDBLastSynced, system.EmployeeDBLastSynced)
	if err != nil {
		slog.Error("error updating system info: " + err.Error())
		return err
	}

	return tx.Commit()
}
