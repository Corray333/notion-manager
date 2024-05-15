package notion

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Corray333/notion-manager/internal/project"
)

type Time struct {
	Properties struct {
		TotalHours struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Number int    `json:"number"`
		} `json:"Всего ч"`
		TaskName struct {
			ID      string `json:"id"`
			Type    string `json:"type"`
			Formula struct {
				Type   string `json:"type"`
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Название задачи"`
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
	ID         string `json:"id"`
	Properties struct {
		Status struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Status struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Color string `json:"color"`
			} `json:"status"`
		} `json:"Статус"`
		ParentTask struct {
			ID       string        `json:"id"`
			Type     string        `json:"type"`
			Relation []interface{} `json:"relation"`
			HasMore  bool          `json:"has_more"`
		} `json:"Родительская задача"`
		Priority struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Select struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Color string `json:"color"`
			} `json:"select"`
		} `json:"Приоритет"`
		Worker struct {
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
		} `json:"Исполнитель"`
		Product struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
			HasMore bool `json:"has_more"`
		} `json:"Продукт"`
		Estimated struct {
			ID     string  `json:"id"`
			Type   string  `json:"type"`
			Number float64 `json:"number"`
		} `json:"Оценка"`
		Deadline struct {
			ID   string `json:"id"`
			Type string `json:"type"`
			Date struct {
				Start    string      `json:"start"`
				End      interface{} `json:"end"`
				TimeZone interface{} `json:"time_zone"`
			} `json:"date"`
		} `json:"Дедлайн"`
		Task struct {
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
		} `json:"Task"`
	} `json:"properties"`
}

type Storage interface {
	GetClientID(internalID string) (string, error)
	GetInternalID(clientID string) (string, error)
}

func GetTasks(store Storage, project project.Project) ([]Task, error) {
	projectID, err := store.GetInternalID(project.ProjectID)
	if err != nil {
		return nil, err
	}

	resp, err := SearchPages(os.Getenv("TASKS_DB"), map[string]interface{}{
		"filter": map[string]interface{}{
			"and": []map[string]interface{}{
				{
					"timestamp": "created_time",
					"created_time": map[string]interface{}{
						"on_or_after": "2024-05-14",
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
	})
	if err != nil {
		return nil, err
	}

	fmt.Println()
	fmt.Println(string(resp))
	fmt.Println()

	tasks := struct {
		Results []Task `json:"results"`
	}{}
	err = json.Unmarshal(resp, &tasks)
	if err != nil {
		return nil, err
	}
	return tasks.Results, nil

}

func (t *Task) Upload(project project.Project) error {
	if len(t.Properties.Worker.People) == 0 {
		return fmt.Errorf("no worker")
	}
	worker, err := GetWorker(project.WorkersDBID, t.Properties.Worker.People[0].ID)
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
						"content": t.Properties.Task.Title[0].PlainText,
					},
				},
			},
		},
		"Статус": map[string]interface{}{
			"status": map[string]interface{}{
				"name": t.Properties.Status.Status.Name,
			},
		},
		"Исполнитель": map[string]interface{}{
			"relation": []map[string]interface{}{
				{
					"id": worker.ID,
				},
			},
		},
		"Приоритет": map[string]interface{}{
			"select": map[string]interface{}{
				"name": t.Properties.Priority.Select.Name,
			},
		},
		"Оценка": map[string]interface{}{
			"number": t.Properties.Estimated.Number,
		},
		"Дедлайн": map[string]interface{}{
			"date": map[string]interface{}{
				"start": t.Properties.Deadline.Date.Start,
				"end":   t.Properties.Deadline.Date.End,
			},
		},
		"Проект": map[string]interface{}{
			"relation": []map[string]interface{}{
				{
					"id": project.ProjectID,
				},
			},
		},
	}

	test, err := CreatePage(project.TasksDBID, req)

	fmt.Println()
	fmt.Println(string(test))
	fmt.Println()

	return err
}
