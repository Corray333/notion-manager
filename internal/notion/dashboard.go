package notion

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/Corray333/notion-manager/internal/project"
)

type Dashboard struct {
	Results []struct {
		ID            string `json:"id"`
		Type          string `json:"type"`
		ChildDatabase struct {
			Title string `json:"title"`
		} `json:"child_database"`
	} `json:"results"`
}

type ProjectRaw struct {
	Results []struct {
		ID         string `json:"id"`
		Properties struct {
			Name struct {
				Title []struct {
					PlainText string `json:"plain_text"`
				} `json:"title"`
			} `json:"Name"`
			Internal struct {
				Relation []struct {
					ID string `json:"id"`
				} `json:"relation"`
			} `json:"internal"`
		} `json:"properties"`
	} `json:"results"`
}

type Clients struct {
	Results []struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"results"`
}

const DashboardsPage = "7a823697e62b467396b8cc45380cc5ad"

func LoadProjects() []*project.Project {
	url := "https://api.notion.com/v1/blocks/" + DashboardsPage + "/children"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("NOTION_SECRET"))
	req.Header.Set("Notion-Version", "2022-06-28")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	res := Clients{}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil
	}
	projects := []*project.Project{}
	for _, client := range res.Results {
		if client.Type == "child_page" {
			projects = append(projects, loadProject(client.ID)...)
		}
	}
	return projects
}

func loadProject(dashboard_id string) []*project.Project { // 1f92aa7a00954137b88117d0c8330b50
	url := "https://api.notion.com/v1/blocks/" + dashboard_id + "/children"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("NOTION_SECRET"))
	req.Header.Set("Notion-Version", "2022-06-28")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	res := Dashboard{}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil
	}
	projects := []*project.Project{}
	tasksDBID := ""
	workersDBID := ""
	timesDBID := ""
	for _, r := range res.Results {
		switch r.ChildDatabase.Title {
		case "Проекты":
			rows, err := SearchPages(r.ID, map[string]interface{}{})
			if err != nil {
				continue
			}
			raw := ProjectRaw{}
			if err := json.Unmarshal(rows, &raw); err != nil {
				continue
			}
			for _, p := range raw.Results {
				if len(p.Properties.Name.Title) == 0 || len(p.Properties.Internal.Relation) == 0 {
					continue
				}
				project := project.Project{}
				project.ProjectsDBID = r.ID
				project.ProjectID = p.ID
				project.Name = p.Properties.Name.Title[0].PlainText
				project.InternalID = p.Properties.Internal.Relation[0].ID
				projects = append(projects, &project)
			}
		case "Задачи":
			tasksDBID = r.ID
		case "Время":
			timesDBID = r.ID
		case "Ставки":
			workersDBID = r.ID
		}
	}

	for _, proj := range projects {
		proj.TasksDBID = tasksDBID
		proj.TimeDBID = timesDBID
		proj.WorkersDBID = workersDBID

	}

	return projects
}
