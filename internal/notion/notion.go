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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func CreatePage(dbid string, properties map[string]interface{}) ([]byte, error) {
	url := "https://api.notion.com/v1/pages"

	data, err := json.Marshal(map[string]interface{}{
		"parent": map[string]interface{}{
			"type":        "database_id",
			"database_id": dbid,
		},
		"icon": map[string]interface{}{
			"type": "external",
			"external": map[string]interface{}{
				"url": "https://previews.dropbox.com/p/thumb/ACSLWstuTUmidfjVhww2pOtk4vcIbna-VEasIXlNS_zNhqslGNaJNff_t_hcEIYmMzNnhVbprja2rHsl5etNR1giA-xRWz1akdAX19opW8giNglNgSAjOVijibYBa2UMNZbkvM2tkp2WG6xBrlIVrEY1wRT6rCQ1FaPG--oHazNJnOnh5-ZmchWhQebii0NuSZvA26UJUcYF-RiryWh2bJWbTuaBK5fE6oBcVze6tDywNepx1wmGLzUwIGPh-Smk72RmO4CyJHjaBROK9FmdIGElmmK0cfqDKqjYWMP8VAshUKhLmLTI06n2X1BSt0255s9pCJudNVutOLvVnQtQszmm/p.png",
			},
		},
		"properties": properties,
	})
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
		return nil, fmt.Errorf("worker not found")
	}
	return &worker.Results[0], nil
}
