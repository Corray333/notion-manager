package notion

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Corray333/notion-manager/internal/project"
)

const TIME_LAYOUT = "2006-01-02T15:04:05.000-07:00"

type Time struct {
	ID         string `json:"id"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor"`
	Properties struct {
		TotalHours struct {
			ID     string  `json:"id"`
			Type   string  `json:"type"`
			Number float64 `json:"number"`
		} `json:"Всего ч"`
		Task struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Задача"`
		WorkDate struct {
			ID   string `json:"id"`
			Type string `json:"type"`
			Date struct {
				Start    string      `json:"start"`
				End      interface{} `json:"end"`
				TimeZone interface{} `json:"time_zone"`
			} `json:"date"`
		} `json:"Дата работ"`
		WhatDid struct {
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
		} `json:"Что делали"`
	}
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

// Task
type Task struct {
	Tries       int    `json:"tries"`
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

type Storage interface {
	GetClientID(internalID string) (string, error)
	GetInternalID(clientID string) (string, error)
	SetClientID(internalID, clientID string) error
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
						"on_or_after": time.Unix(project.LastSynced, 0).Format(TIME_LAYOUT),
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
		fmt.Println("Next cursor applied")
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

func CopyTask(store Storage, project project.Project, taskID string) error {
	resp, err := GetPage(taskID)
	if err != nil {
		return err
	}
	fmt.Println("Response: ", string(resp))
	var task Task
	if err := json.Unmarshal(resp, &task); err != nil {
		return err
	}
	return task.Upload(store, project)

}

var timeMu = &sync.Mutex{}

func (t *Task) Upload(store Storage, project project.Project) error {

	if _, err := store.GetClientID(t.ID); err == nil {
		return nil
	}

	var worker *Worker
	err := error(nil)
	if len(t.Properties.Worker.People) != 0 {
		// TODO: handle errors, except of worker not found
		worker, _ = GetWorker(project.WorkersDBID, t.Properties.Worker.People[0].ID)
	}

	parentTask := []struct {
		ID string `json:"id"`
	}{}

	// Find parent task
	if len(t.Properties.ParentTask.Relation) > 0 && t.Properties.ParentTask.Relation[0].ID != t.ID {
		parentId, err := store.GetClientID(t.Properties.ParentTask.Relation[0].ID)
		if err != nil {
			err = CopyTask(store, project, t.Properties.ParentTask.Relation[0].ID)
			if err != nil {
				return err
			}
			parentId, err = store.GetClientID(t.Properties.ParentTask.Relation[0].ID)
			if err != nil {
				return err
			}
		}
		parentTask = append(parentTask, struct {
			ID string `json:"id"`
		}{
			ID: parentId,
		})
	}

	// TODO: add request constructor

	if len(t.Properties.Task.Title) == 0 {
		fmt.Println()
		fmt.Println("GG:", t)
		fmt.Println()
	}

	req := map[string]interface{}{
		"Name": map[string]interface{}{
			"type": "title",
			"title": []map[string]interface{}{
				{
					"type": "text",
					"text": map[string]interface{}{
						"content": t.Properties.Task.Title[0].PlainText,
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

	// Find icon. It depends on tag, but it is "Иерархическая задача" if it has subtasks
	var icon string
	if len(t.Properties.Tags.MultiSelect) > 0 {
		icon = t.Properties.Tags.MultiSelect[0].Name
	} else if len(t.Properties.Subtasks.Relation) > 0 {
		icon = "Иерархическая задача"
	}

	test, err := CreatePage(project.TasksDBID, req, icon)
	if err != nil {
		return err
	}

	var response struct {
		ID string `json:"id"`
	}

	if err = json.Unmarshal(test, &response); err != nil {
		return err
	}
	if err := store.SetClientID(t.ID, response.ID); err != nil {
		return fmt.Errorf("failed to save task in db: %w", err)
	}

	timeMu.Lock()
	created_at, _ := time.Parse(TIME_LAYOUT, t.CreatedTime)
	if project.LastSynced < created_at.Unix() {
		project.LastSynced = created_at.Unix()
	}
	timeMu.Unlock()

	return err
}

func GetTimes(store Storage, project project.Project) ([]Time, error) {
	projectID, err := store.GetInternalID(project.ProjectID)
	if err != nil {
		return nil, err
	}

	resp, err := SearchPages(os.Getenv("TIME_DB"), map[string]interface{}{
		"filter": map[string]interface{}{
			"and": []map[string]interface{}{
				{
					"timestamp": "created_time",
					"created_time": map[string]interface{}{
						"on_or_after": time.Unix(project.LastSynced, 0).Format(TIME_LAYOUT),
					},
				},
				{
					"property": "Проект",
					"rollup": map[string]interface{}{
						"any": map[string]interface{}{
							"relation": map[string]interface{}{
								"contains": projectID,
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	times := struct {
		Results    []Time `json:"results"`
		HasMore    bool   `json:"has_more"`
		NextCursor string `json:"next_cursor"`
	}{}

	err = json.Unmarshal(resp, &times)
	if err != nil {
		return nil, err
	}
	return times.Results, nil
}

func (t *Time) Upload(store Storage, project project.Project) error {

	task, err := store.GetClientID(t.Properties.Task.Relation[0].ID)
	if err != nil {
		return err
	}

	req := map[string]interface{}{
		"Name": map[string]interface{}{
			"type": "title",
			"title": []map[string]interface{}{
				{
					"type": "text",
					"text": map[string]interface{}{
						"content": t.Properties.WhatDid.Title[0].PlainText,
					},
				},
			},
		},
		"Всего ч": map[string]interface{}{
			"number": t.Properties.TotalHours.Number,
		},
		"Задача": map[string]interface{}{
			"relation": []map[string]interface{}{
				{
					"id": task,
				},
			},
		},
	}

	_, err = CreatePage(project.TimeDBID, req, "")

	return err
}
