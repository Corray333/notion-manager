package notion

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/Corray333/notion-manager/internal/mindmap"
	"github.com/Corray333/notion-manager/pkg/notion"
)

type PageCreated struct {
	ID string `json:"id"`
}

func CreateMindmapTask(projectID string, task *mindmap.Task, parentID string, level int) error {
	fmt.Printf("Creating task: %s\n", task.Title)

	// Основная структура для страницы задачи
	req := map[string]interface{}{
		"Оценка": map[string]interface{}{
			"number": task.Hours,
		},
		"Task": map[string]interface{}{
			"type": "title",
			"title": []map[string]interface{}{
				{
					"type": "text",
					"text": map[string]interface{}{
						"content": task.Title,
					},
				},
			},
		},
	}

	// Если проект указан, связываем задачу с проектом
	if projectID != "" {
		req["Продукт"] = map[string]interface{}{
			"relation": []map[string]interface{}{
				{
					"id": projectID,
				},
			},
		}
	}

	content := []map[string]interface{}{}
	// Если задача второго уровня, добавляем родительскую задачу
	if parentID != "" {
		req["Родительская задача"] = map[string]interface{}{
			"relation": []map[string]interface{}{
				{
					"id": parentID,
				},
			},
		}

		// Массив для вложенных задач, если они есть
		content = CreateCheckboxes(task.Subtasks)
	}

	// Создаем страницу задачи в Notion
	resp, err := notion.CreatePage(os.Getenv("TASKS_DB"), req, content, "")
	if err != nil {
		slog.Error("Notion error while creating task: " + err.Error())
		return err
	}

	var page PageCreated
	if err := json.Unmarshal(resp, &page); err != nil {
		slog.Error("Error unmarshalling response: " + err.Error())
		return err
	}

	if level == 0 {
		// Рекурсивно создаем подзадачи
		for _, subtask := range task.Subtasks {
			if err := CreateMindmapTask(projectID, &subtask, page.ID, level+1); err != nil {
				return err
			}
		}
	}

	return nil
}

func CreateCheckboxes(tasks []mindmap.Task) []map[string]interface{} {
	content := []map[string]interface{}{}

	for _, subtask := range tasks {
		content = append(content, map[string]interface{}{
			"type": "to_do",
			"to_do": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{
						"type": "text",
						"text": map[string]interface{}{
							"content": subtask.Title,
						},
					},
				},
				"checked":  false,
				"children": CreateCheckboxes(subtask.Subtasks),
			},
		})
	}

	return content
}

func CreateMindmapTasks(projectName string, tasks []mindmap.Task) error {
	fmt.Println("Creating tasks for project:", projectName)

	fmt.Println(projectName)
	// Поиск проекта в Notion
	projectFilter := map[string]interface{}{
		"filter": map[string]interface{}{
			"property": "Name",
			"rich_text": map[string]interface{}{
				"contains": projectName,
			},
		},
	}
	projectsResp, err := notion.SearchPages(os.Getenv("PROJECTS_DB"), projectFilter)
	if err != nil {
		slog.Error("Notion error while searching projects: " + err.Error())
		return err
	}

	// Извлекаем ID проекта
	projects := Projects{}
	if err := json.Unmarshal(projectsResp, &projects); err != nil {
		slog.Error("Error unmarshalling projects response: " + err.Error())
		return err
	}

	projectID := ""
	if len(projects.Results) != 0 {
		projectID = projects.Results[0].ID
	}

	// Создаем задачи для проекта
	for _, task := range tasks {
		if err := CreateMindmapTask(projectID, &task, "", 0); err != nil {
			return err
		}
	}

	return nil
}

type Projects struct {
	Results []struct {
		ID string `json:"id"`
	}
}
