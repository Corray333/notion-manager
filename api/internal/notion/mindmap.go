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

func CreateMindmapTask(projectID string, task *mindmap.Task) error {

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
	for _, subpoint := range task.Subpoints {
		content = append(content, map[string]interface{}{
			"type": "to_do",
			"to_do": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{
						"type": "text",
						"text": map[string]interface{}{
							"content": subpoint,
						},
					},
				},
				"checked": false,
			},
		})
	}

	resp, err := notion.CreatePage(os.Getenv("TASKS_DB"), req, content, "")
	if err != nil {
		slog.Error("notion error while creating task: " + err.Error())
		return err
	}

	var page PageCreated
	if err := json.Unmarshal(resp, &page); err != nil {
		slog.Error("error unmarshalling response: " + err.Error())
		return err
	}

	for _, subtask := range task.Subtasks {
		req := map[string]interface{}{
			"Оценка": map[string]interface{}{
				"number": subtask.Hours,
			},
			"Task": map[string]interface{}{
				"type": "title",
				"title": []map[string]interface{}{
					{
						"type": "text",
						"text": map[string]interface{}{
							"content": subtask.Title,
						},
					},
				},
			},
			"Родительская задача": map[string]interface{}{
				"relation": []map[string]interface{}{
					{
						"id": page.ID,
					},
				},
			},
		}

		content := []map[string]interface{}{}
		for _, subpoint := range subtask.Subpoints {
			content = append(content, map[string]interface{}{
				"type": "to_do",
				"to_do": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{
							"type": "text",
							"text": map[string]interface{}{
								"content": subpoint,
							},
						},
					},
					"checked": false,
				},
			})
		}

		_, err := notion.CreatePage(os.Getenv("TASKS_DB"), req, content, "")
		if err != nil {
			slog.Error("notion error while creating task: " + err.Error())
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

func CreateMindmapTasks(projectName string, tasks []mindmap.Task) error {
	fmt.Println("Creating tasks for project:", projectName)
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
		slog.Error("notion error while searching projects: " + err.Error())
		return err
	}

	projects := Projects{}
	if err := json.Unmarshal(projectsResp, &projects); err != nil {
		slog.Error("error unmarshalling projects response: " + err.Error())
		return err
	}

	projectID := ""
	if len(projects.Results) != 0 {
		projectID = projects.Results[0].ID
	}

	for _, task := range tasks {
		if err := CreateMindmapTask(projectID, &task); err != nil {
			return err
		}
	}
	return nil
}
