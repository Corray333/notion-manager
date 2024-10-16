package external

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/Corray333/task_tracker/internal/entities"
	"github.com/Corray333/task_tracker/pkg/notion"
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

func (e *External) GetTimes(lastSynced int64) (times []entities.Time, lastUpdate int64, err error) {
	filter := map[string]interface{}{
		"filter": map[string]interface{}{
			"timestamp": "last_edited_time",
			"last_edited_time": map[string]interface{}{
				"after": time.Unix(lastSynced, 0).Format(notion.TIME_LAYOUT),
			},
		},
		"sorts": []map[string]interface{}{
			{
				"timestamp": "created_time",
				"direction": "ascending",
			},
		},
	}

	resp, err := notion.SearchPages(os.Getenv("TIMES_DB"), filter)
	if err != nil {
		return nil, 0, err
	}
	project := struct {
		Results []Time `json:"results"`
	}{}
	if err := json.Unmarshal(resp, &project); err != nil {
		return nil, 0, err
	}

	times = []entities.Time{}
	for _, w := range project.Results {
		times = append(times, entities.Time{
			Description: w.Properties.WhatDid.Title[0].PlainText,
			ID:          strings.ReplaceAll(w.ID, "-", ""),
		})

		lastEditedTime, err := time.Parse(notion.TIME_LAYOUT_IN, w.LastEditedTime)
		if err != nil {
			return nil, 0, err
		}

		lastUpdate = lastEditedTime.Unix()
	}
	return times, lastUpdate, nil
}
