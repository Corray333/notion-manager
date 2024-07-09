package notion

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Corray333/notion-manager/internal/project"
)

const TIME_LAYOUT = "2006-01-02T15:04:05.000-07:00"
const TIME_LAYOUT_IN = "2006-01-02T15:04:05.999Z07:00"

type TableType string

var IsSyncing = false

const (
	TimeTable    TableType = "time"
	TaskTable    TableType = "task"
	ProjectTable TableType = "project"
)

type Error struct {
	err        error
	project    project.Project
	table_type TableType
	id         string
}

func (e Error) String() string {
	return fmt.Sprintf("Error: %s, Table: %s, Project: %s, ID: %s", e.err.Error(), e.table_type, e.project.Name, e.id)
}

func (err Error) Unpack() (string, string, string, string) {
	return err.project.ProjectID,
		string(err.table_type),
		err.err.Error(),
		err.id
}

// Task
type Storage interface {
	NewProject(proj *project.Project) error
	GetProjects() ([]project.Project, error)
	SetLastSynced(project *project.Project) error
	GetClientID(internalID string) (string, error)
	GetInternalID(clientID string) (string, error)
	SetClientID(internalID, clientID string) error
	SaveRowsToBeUpdated(Validation)
	GetRowsToBeUpdated() ([]Validation, error)
	GetRowsToBeUpdatedByProject(projectID string) ([]Validation, error)
	RemoveRowToBeUpdated(internalID string) error
	SaveError(errSave Error) error
}

type Validation struct {
	Title      string `json:"title" db:"title"`             // Title of page in database
	Type       string `json:"type" db:"type"`               // Type of database
	InternalID string `json:"internal_id" db:"internal_id"` // ID of page in internal dashboard
	ClientID   string `json:"client_id" db:"client_id"`     // ID of page in client dashboard
	Errors     string `json:"errors" db:"errors"`           // Errors encountered while validating
	ProjectID  string `json:"project_id" db:"project_id"`   // ID of project
}

func StartSync(store Storage) error {
	if IsSyncing {
		return errors.New("is already syncing")
	} else {
		IsSyncing = true
	}

	go func() {

		defer func() {
			IsSyncing = false
		}()

		for _, proj := range LoadProjects() {
			store.NewProject(proj)
		}

		projects, err := store.GetProjects()
		if err != nil {
			// return err
		}

		for _, project := range projects {
			project.Schema, _ = GetSchema(project.TasksDBID)
			tasks, err := GetTasks(store, project, "")
			if err != nil {
				store.SaveError(Error{
					err:        errors.Join(errors.New("error while getting tasks: "), err),
					table_type: TaskTable,
					project:    project,
				})
			}
			fmt.Printf("Loaded %d tasks.\n", len(tasks))
			for _, task := range tasks {
				err := task.Upload(store, &project)
				if err != nil {
					fmt.Println(err)
					store.SaveError(Error{
						err:        err,
						table_type: TaskTable,
						project:    project,
						id:         task.ID,
					})
				}
				if err := store.SetLastSynced(&project); err != nil {
					store.SaveError(Error{
						err:        err,
						table_type: ProjectTable,
						project:    project,
					})
				}
			}

			if project.TimeDBID != "" {
				project.Schema, _ = GetSchema(project.TimeDBID)
				times, err := GetTimes(project.TasksLastSynced, project.InternalID, "")
				if err != nil {
					store.SaveError(Error{
						err:        errors.Join(errors.New("error while getting time rows: "), err),
						table_type: TimeTable,
						project:    project,
					})
				}
				fmt.Printf("Loaded %d times.", len(times))
				for _, time := range times {
					if err := time.Upload(store, &project); err != nil {
						store.SaveError(Error{
							err:        err,
							table_type: TaskTable,
							project:    project,
							id:         time.ID,
						})
					}
					if err := store.SetLastSynced(&project); err != nil {
						store.SaveError(Error{
							err:        err,
							table_type: ProjectTable,
							project:    project,
						})
					}
				}
			}
		}
	}()
	return nil
}

func SearchPages(dbid string, filter map[string]interface{}) ([]byte, error) {
	url := "https://api.notion.com/v1/databases/" + dbid + "/query"

	data, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("NOTION_SECRET"))
	req.Header.Set("Notion-Version", "2022-06-28")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func GetPage(pageid string) ([]byte, error) {
	url := "https://api.notion.com/v1/pages/" + pageid

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("NOTION_SECRET"))
	req.Header.Set("Notion-Version", "2022-06-28")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("notion error while getting page: %s", string(body))
	}

	return body, nil
}

func CreatePage(dbid string, properties interface{}, icon string) ([]byte, error) {
	url := "https://api.notion.com/v1/pages"

	reqBody := map[string]interface{}{
		"parent": map[string]interface{}{
			"type":        "database_id",
			"database_id": dbid,
		},
		"properties": properties,
	}
	if icons[icon] != "" {
		reqBody["icon"] = map[string]interface{}{
			"type": "external",
			"external": map[string]interface{}{
				"url": icons[icon],
			},
		}
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("NOTION_SECRET"))
	req.Header.Set("Notion-Version", "2022-06-28")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("notion error %s while creating page with properties %s", string(body), string(data))
	}

	return body, nil
}

func UpdatePage(pageid string, properties interface{}) ([]byte, error) {
	url := "https://api.notion.com/v1/pages/" + pageid

	reqBody := map[string]interface{}{
		"properties": properties,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("NOTION_SECRET"))
	req.Header.Set("Notion-Version", "2022-06-28")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("notion error %s while updating page with properties %s", string(body), string(data))
	}

	return body, nil

}

func FixBroken(store Storage) error {
	projects, err := store.GetProjects()
	if err != nil {
		panic(err)
	}

	for _, project := range projects {
		pages, err := store.GetRowsToBeUpdatedByProject(project.ProjectID)
		if err != nil {
			return err
		}
		for _, page := range pages {
			if page.Type == "task" {
				task, err := GetTask(page.InternalID)
				if err != nil {
					return err
				}
				if err := task.Update(store, &project); err != nil {
					return err
				}
			} else if page.Type == "time" {
				time, err := GetTime(page.InternalID)
				if err != nil {
					return err
				}
				if err := time.Update(store, &project); err != nil {
					return err
				}
			}
		}
	}

	return nil

}

func getWorker(dbid, workerId string) (*Worker, error) {
	filter := map[string]interface{}{
		"filter": map[string]interface{}{
			"property": "Ссылка",
			"people": map[string]interface{}{
				"contains": workerId,
			},
		},
	}

	resp, err := SearchPages(dbid, filter)
	if err != nil {
		return nil, err
	}
	worker := struct {
		Results []Worker `json:"results"`
	}{}
	json.Unmarshal(resp, &worker)
	if len(worker.Results) == 0 {
		return nil, fmt.Errorf("worker not found: %s", string(resp))
	}
	return &worker.Results[0], nil
}

var icons = map[string]string{
	"iOS":     "https://i.postimg.cc/kGZPbxtx/ios.png",
	"Flutter": "https://i.postimg.cc/0QVs8gkX/flutter.png",
	"Android": "https://i.postimg.cc/0NqxndXn/android.png",
	"Иерархическая задача": "https://i.postimg.cc/yYgHzrwn/queen.png",
	"Дизайн":     "https://i.postimg.cc/hjDnW4B7/design.png",
	"Backend":    "https://i.postimg.cc/6QLKm2DX/backend.png",
	"Web":        "https://i.postimg.cc/t4FGsnRx/web.png",
	"Менеджмент": "https://i.postimg.cc/FFn48Qn7/management.png",
	"Sprint":     "https://i.postimg.cc/Qx2rNJDD/Vector.png",
	"Meeting":    "https://i.postimg.cc/brdjr6dy/Group-5272.png",
}

// Worker
type Worker struct {
	ID         string `json:"id"`
	Properties struct {
		Link struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			People []struct {
				Object    string `json:"object"`
				ID        string `json:"id"`
				Name      string `json:"name"`
				AvatarURL string `json:"avatar_url"`
				Type      string `json:"type"`
				Person    struct {
					Email string `json:"email"`
				} `json:"person"`
			} `json:"people"`
		} `json:"Ссылка"`
		Salary struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Number int    `json:"number"`
		} `json:"Ставка в час"`
		Direction struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Select struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Color string `json:"color"`
			} `json:"select"`
		} `json:"Направление"`
		Name struct {
			ID    string `json:"id"`
			Type  string `json:"type"`
			Title []struct {
				Type string `json:"type"`
				Text struct {
					Content string      `json:"content"`
					Link    interface{} `json:"link"`
				} `json:"text"`
				Annotations struct {
					Bold          bool   `json:"bold"`
					Italic        bool   `json:"italic"`
					Strikethrough bool   `json:"strikethrough"`
					Underline     bool   `json:"underline"`
					Code          bool   `json:"code"`
					Color         string `json:"color"`
				} `json:"annotations"`
				PlainText string      `json:"plain_text"`
				Href      interface{} `json:"href"`
			} `json:"title"`
		} `json:"Name"`
	}
}

type Schema struct {
	Properties map[string]interface{} `json:"properties"`
}

func GetSchema(dbid string) ([]string, error) {
	url := "https://api.notion.com/v1/databases/" + dbid

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("NOTION_SECRET"))
	req.Header.Set("Notion-Version", "2022-06-28")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("notion error while getting page: %s", string(body))
	}

	schema := Schema{}
	if err := json.Unmarshal(body, &schema); err != nil {
		return nil, err
	}

	res := []string{}
	for k := range schema.Properties {
		res = append(res, k)
	}
	return res, nil
}
