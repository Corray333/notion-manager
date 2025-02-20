package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
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

func GetHTTPClient() *http.Client {

	proxyURL, _ := url.Parse("http://uLQaWYgF:HTPiw5k5@154.195.127.136:64840")

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	return &http.Client{
		Transport: transport,
	}
}

type searchResponse struct {
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor"`
}

func SearchPages(dbid string, filter map[string]interface{}) ([]byte, error) {
	urlStr := "https://api.notion.com/v1/databases/" + dbid + "/query"

	data, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("NOTION_SECRET"))
	req.Header.Set("Notion-Version", "2022-06-28")
	req.Header.Set("Content-Type", "application/json")

	client := GetHTTPClient()

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
		slog.Error("notion error while searching pages: " + string(body))
		return nil, fmt.Errorf("notion error %s while searching pages with body %s", string(body), string(data))
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

func CreatePage(dbid string, properties interface{}, content interface{}, icon string) ([]byte, error) {
	url := "https://api.notion.com/v1/pages"

	reqBody := map[string]interface{}{
		"parent": map[string]interface{}{
			"type":        "database_id",
			"database_id": dbid,
		},
		"properties": properties,
	}

	if content != nil {
		reqBody["children"] = content
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

	client := GetHTTPClient()
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
