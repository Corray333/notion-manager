package service

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Corray333/task_tracker/internal/entities"
	"github.com/go-co-op/gocron"
)

type repository interface {
	GetEmployees() (employees []entities.Employee, err error)
	GetProjects(userID string) (projects []entities.Project, err error)
	GetTasks(userID string, projectID string) (tasks []entities.Task, err error)
	GetTimesMsg() (times []entities.TimeMsg, err error)

	SetEmployees(employees []entities.Employee) error
	SetTasks(tasks []entities.Task) error
	SetProjects(projects []entities.Project) error
	SaveTimeWriteOf(time *entities.TimeMsg) error

	GetInvalidRows() (times []entities.Row, err error)
	SetInvalidRows(times []entities.Row) error
	MarkInvalidRowsAsSent(times []entities.Row) error

	GetSystemInfo() (*entities.System, error)

	SetSystemInfo(system *entities.System) error
	MarkTimeAsSent(timeID int64) error

	GetEmployeeByID(employeeID string) (employee entities.Employee, err error)
}

type external interface {
	GetEmployees(lastSynced int64) (employees []entities.Employee, lastUpdate int64, err error)
	GetTasks(lastSynced int64, startCursor string) (tasks []entities.Task, lastUpdate int64, err error)
	GetProjects(lastSynced int64) (projects []entities.Project, lastUpdate int64, err error)
	GetTimes(lastSynced int64, startCursor string) (times []entities.Time, lastUpdate int64, err error)

	WriteOfTime(time *entities.TimeMsg) error

	SendNotification(msg []entities.Row) error
}

type Service struct {
	repo     repository
	external external
	cron     *gocron.Scheduler
}

func New(repo repository, external external) *Service {
	loc, _ := time.LoadLocation("Europe/Moscow")
	s := gocron.NewScheduler(loc)

	return &Service{
		repo:     repo,
		external: external,
		cron:     s,
	}
}

func (s *Service) Run() {
	go s.StartUpdatingWorker()
	go s.StartOutboxWorker()

	s.cron.Every(1).Day().At("20:00").Do(s.CheckInvalid)
	s.cron.StartBlocking()
}

func (s *Service) CheckInvalid() {
	rows, err := s.repo.GetInvalidRows()
	if err != nil {
		slog.Error("error getting times: " + err.Error())
		return
	}

	grouped := s.groupByEmployeeID(rows)
	for _, rows := range grouped {
		if err := s.external.SendNotification(rows); err != nil {
			slog.Error("error sending notification: " + err.Error())
			continue
		}

		if err := s.repo.MarkInvalidRowsAsSent(rows); err != nil {
			slog.Error("error marking invalid rows as sent: " + err.Error())
			continue
		}
	}

}

func (s *Service) groupByEmployeeID(rows []entities.Row) map[string][]entities.Row {
	grouped := map[string][]entities.Row{}
	for _, row := range rows {
		if row.Employee == "" && row.EmployeeID != "" {
			employee, err := s.repo.GetEmployeeByID(row.EmployeeID)
			if err == nil {
				row.Employee = employee.Username
			}
		}
		grouped[row.Employee] = append(grouped[row.Employee], row)
	}
	return grouped
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
		times, err := s.repo.GetTimesMsg()
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

	fmt.Println("Getting times")
	times, timesLastUpdate, err := s.external.GetTimes(system.TimesDBLastSynced, "")
	if err != nil {
		return false, err
	}

	invalidTimes := s.ValidateTimes(times)

	if len(invalidTimes) > 0 {
		if err := s.repo.SetInvalidRows(invalidTimes); err != nil {
			return false, err
		}
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

	invalidTasks := s.ValidateTasks(tasks)
	if len(invalidTasks) > 0 {
		if err := s.repo.SetInvalidRows(invalidTasks); err != nil {
			return false, err
		}
	}

	if err := s.repo.SetTasks(tasks); err != nil {
		return false, err
	}

	system.EmployeeDBLastSynced = employeesLastUpdate
	system.ProjectsDBLastSynced = projectsLastUpdate
	system.TasksDBLastSynced = tasksLastUpdate
	system.TimesDBLastSynced = timesLastUpdate

	if err := s.repo.SetSystemInfo(system); err != nil {
		return false, err
	}

	return len(employees) > 0 || len(projects) > 0 || len(tasks) > 0, nil
}

func (s *Service) WriteOfTime(time *entities.TimeMsg) error {
	return s.repo.SaveTimeWriteOf(time)
}

var forbiddenWords = []string{
	"Фикс",
	"Пофиксить",
	"Фиксить",
	"Правка",
	"Править",
	"Поправить",
	"Исправить",
	"Правки",
	"Исправление",
	"Баг",
	"Безуспешно",
	"Разобраться",
}

func containsForbiddenWord(input string) (string, bool) {
	lowerInput := strings.ToLower(input)
	for _, word := range forbiddenWords {
		if strings.Contains(lowerInput, strings.ToLower(word)) {
			return word, true
		}
	}
	return "", false
}

// TODO: replace with outbox pattern
func (s *Service) ValidateTimes(times []entities.Time) []entities.Row {
	invalidTimes := []entities.Row{}
	for _, time := range times {
		if word, contains := containsForbiddenWord(time.Description); contains {
			time.Description = strings.ReplaceAll(time.Description, strings.ToLower(word), "<b><i>"+strings.ToLower(word)+"</i></b>")
			time.Description = strings.ReplaceAll(time.Description, word, "<b><i>"+word+"</i></b>")
			invalidTimes = append(invalidTimes, time.ToRow())
		}
	}
	return invalidTimes
}

func (s *Service) ValidateTasks(tasks []entities.Task) []entities.Row {
	invalidTasks := []entities.Row{}
	for _, task := range tasks {
		if word, contains := containsForbiddenWord(task.Title); contains {
			task.Title = strings.ReplaceAll(task.Title, strings.ToLower(word), "<b><i>"+strings.ToLower(word)+"</i></b>")
			task.Title = strings.ReplaceAll(task.Title, word, "<b><i>"+word+"</i></b>")
			invalidTasks = append(invalidTasks, task.ToRow())
		}
	}

	return invalidTasks
}
