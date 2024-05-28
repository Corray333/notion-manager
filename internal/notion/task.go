package notion

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Corray333/notion-manager/internal/project"
)

var (
	ErrTaskNoTitle    = "task title is empty"
	ErrTaskNoWorker   = "task worker is empty"
	ErrTaskNoProduct  = "task product is empty"
	ErrTaskNoDeadline = "task deadline is empty"
)

type Task struct {
	ID          string `json:"id"`
	CreatedTime string `json:"created_time"`
	Properties  struct {
		Tags struct {
			MultiSelect []struct {
				Name string `json:"name"`
			} `json:"multi_select"`
		} `json:"Теги"`
		Status struct {
			Status struct {
				Name string `json:"name"`
			} `json:"status"`
		} `json:"Статус"`
		ParentTask struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Родительская задача"`
		Priority struct {
			Select struct {
				Name string `json:"name"`
			} `json:"select"`
		} `json:"Приоритет"`
		Worker struct {
			People []struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			} `json:"people"`
		} `json:"Исполнитель"`
		Product struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Продукт"`
		Estimated struct {
			Number float64 `json:"number"`
		} `json:"Оценка"`
		Subtasks struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Подзадачи"`
		Deadline struct {
			Date struct {
				Start    string      `json:"start"`
				End      interface{} `json:"end"`
				TimeZone interface{} `json:"time_zone"`
			} `json:"date"`
		} `json:"Дедлайн"`
		Task struct {
			Title []struct {
				PlainText string `json:"plain_text"`
			} `json:"title"`
		} `json:"Task"`
	} `json:"properties"`
}

func (t *Task) Validate(store Storage, client_id string) {
	errs := ""
	title := ""
	if t.Properties.Task.Title == nil || len(t.Properties.Task.Title) == 0 {
		errs += ErrTaskNoTitle + ", "
	} else {
		title = t.Properties.Task.Title[0].PlainText
	}
	if t.Properties.Product.Relation == nil || len(t.Properties.Product.Relation) == 0 {
		errs += ErrTaskNoProduct + ", "
	}
	if t.Properties.Worker.People == nil || len(t.Properties.Worker.People) == 0 {
		errs += ErrTaskNoWorker + ", "
	}
	if t.Properties.Deadline.Date.Start == "" {
		errs += ErrTaskNoDeadline + ", "
	}

	if len(errs) > 0 {
		fmt.Println("Validation failed: ", t.ID, " --- ", client_id, " --- ", errs)
		store.SaveRowsToBeUpdated(Validation{
			ClientID:   client_id,
			InternalID: t.ID,
			Title:      title,
			Errors:     errs,
			Type:       "task",
		})
	}
}

type GetTasksResponse struct {
	Results    []Task `json:"results"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor"`
}

func GetTasks(store Storage, project project.Project, cursor string) ([]Task, error) {
	projectID, err := store.GetInternalID(project.ProjectID)
	if err != nil {
		return nil, err
	}
	req := map[string]interface{}{
		"filter": map[string]interface{}{
			"and": []map[string]interface{}{
				{
					"timestamp": "created_time",
					"created_time": map[string]interface{}{
						"after": time.Unix(project.TasksLastSynced, 0).Format(TIME_LAYOUT),
					},
				},
				{
					"property": "Продукт",
					"relation": map[string]interface{}{
						"contains": projectID,
					},
				},
			},
		},
		"sorts": []map[string]interface{}{
			{
				"timestamp": "created_time",
				"direction": "ascending",
			},
		},
	}

	if cursor != "" {
		req["start_cursor"] = cursor
	}

	resp, err := SearchPages(os.Getenv("TASKS_DB"), req)
	if err != nil {
		return nil, err
	}

	tasks := GetTasksResponse{}

	err = json.Unmarshal(resp, &tasks)
	if err != nil {
		return nil, err
	}

	if tasks.HasMore {
		moreTasks, err := GetTasks(store, project, tasks.NextCursor)
		if err != nil {
			return nil, err
		}
		return append(tasks.Results, moreTasks...), nil
	}

	return tasks.Results, nil

}

func (t *Task) ConstructRequest(store Storage, project *project.Project) (map[string]interface{}, error) {
	var worker *Worker
	if len(t.Properties.Worker.People) != 0 {
		// TODO: handle errors
		worker, _ = getWorker(project.WorkersDBID, t.Properties.Worker.People[0].ID)
	}

	parentTask := []struct {
		ID string `json:"id"`
	}{}

	// Find parent task
	if len(t.Properties.ParentTask.Relation) > 0 && t.Properties.ParentTask.Relation[0].ID != t.ID {
		parentId, err := store.GetClientID(t.Properties.ParentTask.Relation[0].ID)
		if err != nil {
			if err != sql.ErrNoRows {
				return nil, err
			}
			resp, err := GetPage(t.Properties.ParentTask.Relation[0].ID)
			if err != nil {
				return nil, err
			}
			var task Task
			if err := json.Unmarshal(resp, &task); err != nil {
				return nil, err
			}

			if err := task.Upload(store, project); err != nil {
				return nil, err
			}

			parentId, err = store.GetClientID(t.Properties.ParentTask.Relation[0].ID)
			if err != nil {
				return nil, err
			}
		}
		parentTask = append(parentTask, struct {
			ID string `json:"id"`
		}{
			ID: parentId,
		})
	}

	// TODO: add request constructorПип
	title := ""
	for _, t := range t.Properties.Task.Title {
		title += t.PlainText
	}

	req := map[string]interface{}{
		"Name": map[string]interface{}{
			"type": "title",
			"title": []map[string]interface{}{
				{
					"type": "text",
					"text": map[string]interface{}{
						"content": title,
					},
				},
			},
		},
		"Оценка": map[string]interface{}{
			"number": t.Properties.Estimated.Number,
		},
		"Проект": map[string]interface{}{
			"relation": []map[string]interface{}{
				{
					"id": project.ProjectID,
				},
			},
		},
	}

	if t.Properties.Status.Status.Name != "" {
		req["Статус"] = map[string]interface{}{
			"status": map[string]interface{}{
				"name": t.Properties.Status.Status.Name,
			},
		}
	}

	if t.Properties.Priority.Select.Name != "" {
		req["Приоритет"] = map[string]interface{}{
			"select": map[string]interface{}{
				"name": t.Properties.Priority.Select.Name,
			},
		}
	}

	if len(parentTask) > 0 && parentTask[0].ID != "" {
		req["Родительская задача"] = map[string]interface{}{
			"relation": parentTask,
		}
	}

	deadline := map[string]interface{}{
		"date": map[string]interface{}{},
	}
	if t.Properties.Deadline.Date.End != nil {
		deadline["date"].(map[string]interface{})["end"] = t.Properties.Deadline.Date.End
	}
	if t.Properties.Deadline.Date.Start != "" {
		deadline["date"].(map[string]interface{})["start"] = t.Properties.Deadline.Date.Start
		req["Дедлайн"] = deadline
	}

	if worker != nil {
		req["Исполнитель"] = map[string]interface{}{
			"relation": []map[string]interface{}{
				{
					"id": worker.ID,
				},
			},
		}
	}
	return req, nil

}

func (t *Task) Upload(store Storage, project *project.Project) error {
	if _, err := store.GetClientID(t.ID); err != sql.ErrNoRows {
		return nil
	}

	req, err := t.ConstructRequest(store, project)
	if err != nil {
		return nil
	}

	// Find icon. It depends on tag, but it is "Иерархическая задача" if it has subtasks
	var icon string
	if len(t.Properties.Tags.MultiSelect) > 0 {
		icon = t.Properties.Tags.MultiSelect[0].Name
	} else if len(t.Properties.Subtasks.Relation) > 0 {
		icon = "Иерархическая задача"
	}

	resp, err := CreatePage(project.TasksDBID, req, icon)
	if err != nil {
		return err
	}

	var response struct {
		ID string `json:"id"`
	}
	err = json.Unmarshal(resp, &response)

	if err != nil {
		return err
	}

	if err := store.SetClientID(t.ID, response.ID); err != nil {
		return fmt.Errorf("failed to save task in db: %w", err)
	}
	created_at, _ := time.Parse(TIME_LAYOUT_IN, t.CreatedTime)
	if project.TasksLastSynced < created_at.Unix() {
		project.TasksLastSynced = created_at.Unix()
	}

	t.Validate(store, response.ID)

	return err
}
