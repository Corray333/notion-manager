package entities

import "fmt"

type MsgCreator interface {
	ToMsg() string
}

type RowCreator interface {
	ToRow() Row
}

type Row struct {
	ID          string `json:"id" db:"id"`
	Description string `json:"description" db:"description"`
	Employee    string `json:"employee" db:"employee"`
	EmployeeID  string `json:"employeeID" db:"employee_id"`
}

// TimeMsg godoc
// @Description Represents a time
type TimeMsg struct {
	ID          int64  `json:"id" db:"time_id" example:"0"`
	TaskID      string `json:"taskID" db:"task_id" example:"9eb9de5f-2341-44c6-aae8-fc917394092b"`
	EmployeeID  string `json:"employeeID" db:"employee_id" example:"353198d1-1a40-4b4b-9841-66e7de4de6ea"`
	Duration    int    `json:"duration" db:"duration" example:"1800"`
	Description string `json:"description" db:"description" example:"Мыла попу"`
}

// Task godoc
// @Description Represents a task in the system
type Task struct {
	ID         string `json:"id" db:"task_id" example:"9eb9de5f-2341-44c6-aae8-fc917394092b"`
	Title      string `json:"title" db:"title" example:"Доделать прототип тайм трекера"`
	Status     string `json:"status" db:"status" example:"В работе"`
	ProjectID  string `json:"projectID" db:"project_id" example:"268c4871-39fd-4c78-9681-4d62ae34dcee"`
	EmployeeID string `json:"employeeID" db:"employee_id" example:"353198d1-1a40-4b4b-9841-66e7de4de6ea"`
	Employee   string `json:"employee" example:"Mark"`
}

func (t Task) ToMsg() string {
	return fmt.Sprintf("Ошибка в задаче: [%s](%s)", t.Title, "notion.so/"+t.ID)
}

func (t Task) ToRow() Row {
	return Row{
		ID:          t.ID,
		Description: t.Title,
		Employee:    t.Employee,
		EmployeeID:  t.EmployeeID,
	}
}

type Project struct {
	ID       string `json:"id" db:"project_id" example:"1114675b-93d2-4d67-ad0c-8851b6134af2"`
	Name     string `json:"name" db:"name" example:"Behance"`
	Icon     string `json:"icon" db:"icon" example:"https://prod-files-secure.s3.us-west-2.amazonaws.com/9a2e0635-b9d4-4178-a529-cf6b3bdce29d/7d460da2-42b7-4d5b-8d31-97a327675bc4/behance-1.svg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=AKIAT73L2G45HZZMZUHI%2F20241014%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20241014T055949Z&X-Amz-Expires=3600&X-Amz-Signature=c67998b0c68723e6efb6268baf917f6ae9e4902238a2b146cb054a6cda51c7cf&X-Amz-SignedHeaders=host&x-id=GetObject"`
	IconType string `json:"iconType" db:"icon_type" example:"file"`
	Status   string `json:"status" db:"status" example:"В работе"`
}

type Employee struct {
	ID       string `json:"id" db:"employee_id" example:"790bdb23-c2d3-4154-8497-2ef5f1e6d2ad"`
	Username string `json:"username" db:"username" example:"Mark"`
	Icon     string `json:"icon" db:"icon" example:"https://prod-files-secure.s3.us-west-2.amazonaws.com/9a2e0635-b9d4-4178-a529-cf6b3bdce29d/f2f425d1-efde-46ee-a724-78dcd401bff0/Frame_3.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=AKIAT73L2G45HZZMZUHI%2F20241014%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20241014T062630Z&X-Amz-Expires=3600&X-Amz-Signature=195ddfb2599f4d4e6162d1e467966af275d2bad346414fdb574f61049757e40f&X-Amz-SignedHeaders=host&x-id=GetObject"`
	Email    string `json:"email" db:"email" example:"s0177180@edu.kubsu.ru"`
	Profile  string `json:"profile" db:"profile"`
}

type System struct {
	ID                   int   `json:"id" db:"id"`
	ProjectsDBLastSynced int64 `json:"projectsDBLastSynced" db:"projects_db_last_sync"`
	TasksDBLastSynced    int64 `json:"tasksDBLastSynced" db:"tasks_db_last_sync"`
	EmployeeDBLastSynced int64 `json:"employeeDBLastSynced" db:"employee_db_last_sync"`
	TimesDBLastSynced    int64 `json:"timesDBLastSynced" db:"times_db_last_sync"`
}

type Time struct {
	ID          string `json:"id" db:"time_id" example:"790bdb23-c2d3-4154-8497-2ef5f1e6d2ad"`
	Description string `json:"description" db:"description" example:"Мыла попу"`
	EmployeeID  string `json:"employeeID" db:"employee_id" example:"353198d1-1a40-4b4b-9841-66e7de4de6ea"`
	Employee    string `json:"employee" db:"employee" example:"Mark"`
}

func (t Time) ToMsg() string {
	return fmt.Sprintf("Ошибка в записи времени: [%s](%s)", t.Description, "notion.so/"+t.ID)
}

func (t Time) ToRow() Row {
	return Row{
		ID:          t.ID,
		Description: t.Description,
		Employee:    t.Employee,
		EmployeeID:  t.EmployeeID,
	}
}
