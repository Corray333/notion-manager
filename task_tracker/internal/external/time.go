package external

import (
	"encoding/json"
	"fmt"
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
				ID   string `json:"id"`
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

func (e *External) GetTimes(lastSynced int64, startCursor string) (times []entities.Time, lastUpdate int64, err error) {
	filter := map[string]interface{}{
		"filter": map[string]interface{}{
			"timestamp": "last_edited_time",
			"last_edited_time": map[string]interface{}{
				"after": time.Unix(lastSynced, 0).Format(notion.TIME_LAYOUT),
			},
		},
		"sorts": []map[string]interface{}{
			{
				"timestamp": "last_edited_time",
				"direction": "ascending",
			},
		},
	}

	if startCursor != "" {
		filter["start_cursor"] = startCursor
	}
	lastUpdate = 0

	resp, err := notion.SearchPages(os.Getenv("TIMES_DB"), filter)
	if err != nil {
		return nil, 0, err
	}
	timeResults := struct {
		Results    []Time `json:"results"`
		HasMore    bool   `json:"has_more"`
		NextCursor string `json:"next_cursor"`
	}{}
	if err := json.Unmarshal(resp, &timeResults); err != nil {
		return nil, 0, err
	}

	times = []entities.Time{}
	for _, w := range timeResults.Results {
		times = append(times, entities.Time{
			Description: func() string {
				if len(w.Properties.WhatDid.Title) == 0 {
					return ""
				}

				return w.Properties.WhatDid.Title[0].PlainText
			}(),
			ID: strings.ReplaceAll(w.ID, "-", ""),
			Employee: func() string {
				if len(w.Properties.WhoDid.People) == 0 {
					return ""
				}
				return w.Properties.WhoDid.People[0].Name
			}(),
			EmployeeID: func() string {
				if len(w.Properties.WhoDid.People) == 0 {
					return ""
				}
				return w.Properties.WhoDid.People[0].ID
			}(),
		})

		lastEditedTime, err := time.Parse(notion.TIME_LAYOUT_IN, w.LastEditedTime)
		if err != nil {
			return nil, 0, err
		}

		lastUpdate = lastEditedTime.Unix()
	}

	fmt.Println(lastUpdate)

	if timeResults.HasMore {
		fmt.Println("time has more")
		nextTasks, lastEditedTime, err := e.GetTimes(lastSynced, timeResults.NextCursor)
		if err != nil {
			return nil, 0, err
		}
		lastUpdate = lastEditedTime
		times = append(times, nextTasks...)
	}

	return times, lastUpdate, nil
}
