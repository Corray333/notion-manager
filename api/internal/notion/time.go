package notion

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Corray333/notion-manager/internal/project"
)

var (
	ErrTimeNoTitle      = "time created, but title is empty"
	ErrTimeNoTask       = "time created, but task is empty"
	ErrTimeNoTotalHours = "time created, but total hours is empty"
)

type Time struct {
	ID             string `json:"id"`
	CreatedTime    string `json:"created_time"`
	LastEditedTime string `json:"last_edited_time"`
	Properties     struct {
		TotalHours struct {
			Number float64 `json:"number"`
		} `json:"Всего ч"`
		Analytics struct {
			Relation []struct{} `json:"relation"`
		} `json:"Аналитика"`
		PayableHours struct {
			Formula struct {
				Number float64 `json:"number"`
			} `json:"formula"`
		} `json:"К оплате ч."`
		Task struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Задача"`
		Direction struct {
			Select struct {
				Name string `json:"name"`
			} `json:"select"`
		} `json:"Направление"`
		TaskName struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Название задачи"`
		WorkDate struct {
			Date struct {
				Start    string      `json:"start"`
				End      interface{} `json:"end"`
				TimeZone interface{} `json:"time_zone"`
			} `json:"date"`
		} `json:"Дата работ"`
		WhoDid struct {
			People []struct {
				Name string `json:"name"`
			} `json:"people"`
		} `json:"Исполнитель"`
		EstimateHours struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Оценка ч"`
		CreatedTimeField struct {
			CreatedTime string `json:"created_time"`
		} `json:"Created time"`
		Payment struct {
			Checkbox bool `json:"checkbox"`
		} `json:"Оплата"`
		Project struct {
			Rollup struct {
				Array []struct {
					Relation []struct {
						ID string `json:"id"`
					} `json:"relation"`
				} `json:"array"`
			} `json:"rollup"`
		} `json:"Проект"`
		StatusHours struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Статус ч"`
		Month struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Месяц"`
		ProjectName struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Имя проекта"`
		ProjectStatus struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Статус проекта"`
		WhatDid struct {
			Title []struct {
				PlainText string `json:"plain_text"`
			} `json:"title"`
		} `json:"Что делали"`
		BH struct {
			Formula struct {
				Number float64 `json:"number"`
			} `json:"formula"`
		} `json:"BH"`
		SH struct {
			Number float64 `json:"number"` // Number or null
		} `json:"SH"`
		DH struct {
			Number float64 `json:"number"` // Number or null
		} `json:"DH"`
		BHGS struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"BHGS"`
		WeekNumber struct {
			Formula struct {
				Number float64 `json:"number"`
			} `json:"formula"`
		} `json:"Номер недели"`
		DayNumber struct {
			Formula struct {
				Number float64 `json:"number"`
			} `json:"formula"`
		} `json:"Номер дня"`
		MonthNumber struct {
			Formula struct {
				Number float64 `json:"number"`
			} `json:"formula"`
		} `json:"Номер месяца"`
	} `json:"properties"`
	URL string `json:"url"`
}

func (t *Time) Validate(store Storage, client_id string, project_id string) {
	errs := ""
	title := ""
	if len(t.Properties.WhatDid.Title) == 0 {
		errs += ErrTimeNoTitle + ", "
	} else {
		for _, t := range t.Properties.WhatDid.Title {
			title += t.PlainText
		}
	}
	if len(t.Properties.Task.Relation) == 0 {
		errs += ErrTimeNoTask + ", "
	}
	if t.Properties.TotalHours.Number == 0 {
		errs += ErrTimeNoTotalHours + ", "
	}
	if len(errs) > 0 {
		errs = errs[:len(errs)-2]
		store.SaveRowsToBeUpdated(Validation{
			Type:       "time",
			InternalID: t.ID,
			ClientID:   client_id,
			Title:      title,
			Errors:     errs,
			ProjectID:  project_id,
		})
	} else {
		store.RemoveRowToBeUpdated(t.ID)
	}
}

func GetTimes(lastSynced int64, projectID string, cursor string) ([]Time, error) {
	req := map[string]interface{}{}

	// TODO: refactor
	if projectID != "" {
		req = map[string]interface{}{
			"filter": map[string]interface{}{
				"and": []map[string]interface{}{
					{
						"timestamp": "last_edited_time",
						"last_edited_time": map[string]interface{}{
							"after": time.Unix(lastSynced, 0).Format(TIME_LAYOUT_IN),
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
			"sorts": []map[string]interface{}{
				{
					"timestamp": "created_time",
					"direction": "ascending",
				},
			},
		}
	} else {
		req = map[string]interface{}{
			"filter": map[string]interface{}{
				"timestamp": "last_edited_time",
				"last_edited_time": map[string]interface{}{
					"after": time.Unix(lastSynced, 0).Format(TIME_LAYOUT),
				},
			},
			"sorts": []map[string]interface{}{
				{
					"timestamp": "created_time",
					"direction": "ascending",
				},
			},
		}
	}

	if cursor != "" {
		fmt.Println("Next cursor applied")
		req["start_cursor"] = cursor
	}

	resp, err := SearchPages(os.Getenv("TIME_DB"), req)
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

	if times.HasMore {
		moreTimes, err := GetTimes(lastSynced, projectID, times.NextCursor)
		if err != nil {
			return nil, err
		}
		return append(times.Results, moreTimes...), nil
	}

	return times.Results, nil
}

func GetTime(id string) (Time, error) {
	resp, err := GetPage(id)
	if err != nil {
		return Time{}, err
	}

	time := Time{}
	err = json.Unmarshal(resp, &time)
	if err != nil {
		return Time{}, err
	}

	return time, nil
}

func (t *Time) ConstructRequest(store Storage) (map[string]interface{}, error) {
	if len(t.Properties.Task.Relation) == 0 {
		return nil, errors.New("time has no task, time_id = " + t.ID)
	}
	task, err := store.GetClientID(t.Properties.Task.Relation[0].ID)
	if err != nil {
		return nil, fmt.Errorf("task %s is not copied yet: %w", t.Properties.Task.Relation[0].ID, err)
	}

	req := map[string]interface{}{
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
	title := ""

	for _, t := range t.Properties.WhatDid.Title {
		title += t.PlainText
	}

	if len(t.Properties.WhatDid.Title) > 0 {
		req["Name"] = map[string]interface{}{
			"type": "title",
			"title": []map[string]interface{}{
				{
					"type": "text",
					"text": map[string]interface{}{
						"content": title,
					},
				},
			},
		}
		return req, nil
	} else {
		return req, errors.New(ErrTimeNoTitle)
	}
}

func (t *Time) Upload(store Storage, project *project.Project) error {

	if _, err := store.GetClientID(t.ID); err != nil && err != sql.ErrNoRows {
		return err
	} else if err == nil {
		return t.Update(store, project)
	}

	req, construct_err := t.ConstructRequest(store)
	if construct_err != nil && construct_err.Error() != ErrTimeNoTitle {
		return construct_err
	}

	body, err := CreatePage(project.TimeDBID, req, "")
	if err != nil {
		return err
	}

	resp := struct {
		ID string `json:"id"`
	}{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	if err := store.SetClientID(t.ID, resp.ID); err != nil {
		return fmt.Errorf("failed to save task in db: %w", err)
	}
	created_at, err := time.Parse(TIME_LAYOUT_IN, t.CreatedTime)
	if err != nil {
		return fmt.Errorf("time %s has wrong created time format: %w", t.ID, err)
	}
	if project.TimeLastSynced < created_at.Unix() {
		project.TimeLastSynced = created_at.Unix()
		store.SetLastSynced(project)
	}

	t.Validate(store, resp.ID, project.ProjectID)

	return construct_err
}

func (t *Time) Update(store Storage, project *project.Project) error {
	clientID, err := store.GetClientID(t.ID)
	if err != nil {
		return err
	}
	req, err := t.ConstructRequest(store)
	if err != nil {
		return err
	}

	if _, err := UpdatePage(clientID, req); err != nil {
		return err
	}

	created_at, err := time.Parse(TIME_LAYOUT_IN, t.CreatedTime)
	if err != nil {
		return fmt.Errorf("time %s has wrong created time format: %w", t.ID, err)
	}
	if project.TimeLastSynced < created_at.Unix() {
		project.TimeLastSynced = created_at.Unix()
		store.SetLastSynced(project)
	}

	t.Validate(store, clientID, project.ProjectID)

	return nil
}
