package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

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

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("notion error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
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

	return body, nil
}

func GetWorker(dbid, workerId string) (*Worker, error) {
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
