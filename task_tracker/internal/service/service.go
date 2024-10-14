package service

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Corray333/task_tracker/internal/entities"
)

type repository interface {
	GetEmployees() (employees []entities.Employee, err error)
	GetProjects(userID string) (projects []entities.Project, err error)
	GetTasks(userID string, projectID string) (tasks []entities.Task, err error)
	GetTimes() (times []entities.TimeMsg, err error)

	SetEmployees(employees []entities.Employee) error
	SetTasks(tasks []entities.Task) error
	SetProjects(projects []entities.Project) error
	SaveTimeWriteOf(time *entities.TimeMsg) error

	GetSystemInfo() (*entities.System, error)

	SetSystemInfo(system *entities.System) error
	MarkTimeAsSent(timeID int64) error
}

type external interface {
	GetEmployees(lastSynced int64) (employees []entities.Employee, lastUpdate int64, err error)
	GetTasks(lastSynced int64, startCursor string) (tasks []entities.Task, lastUpdate int64, err error)
	GetProjects(lastSynced int64) (projects []entities.Project, lastUpdate int64, err error)

	WriteOfTime(time *entities.TimeMsg) error
}

type Service struct {
	repo     repository
	external external
}

func New(repo repository, external external) *Service {
	return &Service{
		repo:     repo,
		external: external,
	}
}

func (s *Service) StartUpdatingWorker() {
	for {
		_, err := s.Actualize()
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Minute)
	}
}

func (s *Service) StartOutboxWorker() {
	for {
		slog.Info("Reading outbox")
		times, err := s.repo.GetTimes()
		if err != nil {
			slog.Error("error getting times: " + err.Error())
			continue
		}
		for _, time := range times {
			if err := s.external.WriteOfTime(&time); err != nil {
				slog.Error("error sending time to notion: " + err.Error())
				continue
			}

			// TODO: maybe add compensation of notion query
			if err := s.repo.MarkTimeAsSent(time.ID); err != nil {
				slog.Error("error marking time as sent: " + err.Error())
				continue
			}
		}

		time.Sleep(time.Minute)
	}
}

func (s *Service) GetUsers() ([]entities.Employee, error) {
	return s.repo.GetEmployees()
}

func (s *Service) GetProjects(userID string) ([]entities.Project, error) {
	return s.repo.GetProjects(userID)
}

func (s *Service) GetTasks(userID, projectID string) ([]entities.Task, error) {
	return s.repo.GetTasks(userID, projectID)
}

func (s *Service) Actualize() (updated bool, err error) {
	system, err := s.repo.GetSystemInfo()
	if err != nil {
		return false, err
	}

	fmt.Println("Getting employees")
	employees, employeesLastUpdate, err := s.external.GetEmployees(system.EmployeeDBLastSynced)
	if err != nil {
		return false, err
	}
	if err := s.repo.SetEmployees(employees); err != nil {
		return false, err
	}

	fmt.Println("Getting projects")
	projects, projectsLastUpdate, err := s.external.GetProjects(system.ProjectsDBLastSynced)
	if err != nil {
		return false, err
	}
	if err := s.repo.SetProjects(projects); err != nil {
		return false, err
	}

	fmt.Println("Getting tasks")
	tasks, tasksLastUpdate, err := s.external.GetTasks(system.TasksDBLastSynced, "")
	if err != nil {
		return false, err
	}

	if err := s.repo.SetTasks(tasks); err != nil {
		return false, err
	}

	system.EmployeeDBLastSynced = employeesLastUpdate
	system.ProjectsDBLastSynced = projectsLastUpdate
	system.TasksDBLastSynced = tasksLastUpdate

	s.repo.SetSystemInfo(system)

	return len(employees) > 0 || len(projects) > 0 || len(tasks) > 0, nil
}

func (s *Service) WriteOfTime(time *entities.TimeMsg) error {
	return s.repo.SaveTimeWriteOf(time)
}
