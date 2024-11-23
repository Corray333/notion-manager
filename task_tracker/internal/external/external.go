package external

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/Corray333/task_tracker/internal/entities"
	"github.com/Corray333/task_tracker/pkg/notion"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type External struct {
	tg *TelegramClient
}

type TelegramClient struct {
	bot *tgbotapi.BotAPI
}

func (t *TelegramClient) GetBot() *tgbotapi.BotAPI {
	return t.bot
}

func NewClient(token string) *TelegramClient {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("failed to create bot: ", err)
	}

	bot.Debug = true

	return &TelegramClient{
		bot: bot,
	}
}

func New() *External {
	return &External{
		tg: NewClient(os.Getenv("BOT_TOKEN")),
	}
}

// Worker
type Worker struct {
	ID             string `json:"id"`
	LastEditedTime string `json:"last_edited_time"`
	Icon           struct {
		Type     string   `json:"type"`
		External external `json:"external"`
		File     file     `json:"file"`
		Emoji    emoji    `json:"emoji"`
	} `json:"icon"`
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
		} `json:"Имя"`
	}
}

func (e *External) GetEmployees(lastSynced int64) (employees []entities.Employee, lastUpdate int64, err error) {
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

	resp, err := notion.SearchPages(os.Getenv("EMPLOYEES_DB"), filter)
	if err != nil {
		return nil, 0, err
	}
	worker := struct {
		Results []Worker `json:"results"`
	}{}

	if err := json.Unmarshal(resp, &worker); err != nil {
		return nil, 0, err
	}

	lastUpdate = lastSynced

	employees = []entities.Employee{}
	for _, w := range worker.Results {
		employees = append(employees, entities.Employee{
			ID: func() string {
				if len(w.Properties.Link.People) == 0 {
					return ""
				} else {
					return w.Properties.Link.People[0].ID
				}
			}(),
			Username: func() string {
				if len(w.Properties.Name.Title) == 0 {
					return ""
				}
				return w.Properties.Name.Title[0].PlainText
			}(),
			Icon: func() string {
				if len(w.Properties.Link.People) > 0 {
					return w.Properties.Link.People[0].AvatarURL
				}
				return ""
			}(),
			Email: func() string {
				if len(w.Properties.Link.People) == 0 {
					return ""
				}
				return w.Properties.Link.People[0].Person.Email
			}(),
		})

		lastEditedTime, err := time.Parse(notion.TIME_LAYOUT_IN, w.LastEditedTime)
		if err != nil {
			return nil, 0, err
		}

		lastUpdate = lastEditedTime.Unix()
	}
	return employees, lastUpdate, nil
}

type Task struct {
	ID             string `json:"id"`
	LastEditedTime string `json:"last_edited_time"`
	Properties     struct {
		Tags struct {
			MultiSelect []struct {
				Name string `json:"name"`
			} `json:"multi_select"`
		} `json:"Теги"`
		Status struct {
			Status struct {
				Name string `json:"name"`
			} `json:"status"`
		} `json:"Статус"`
		ParentTask struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Родительская задача"`
		Priority struct {
			Select struct {
				Name string `json:"name"`
			} `json:"select"`
		} `json:"Приоритет"`
		Worker struct {
			People []struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			} `json:"people"`
		} `json:"Исполнитель"`
		Product struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Продукт"`
		Estimated struct {
			Number float64 `json:"number"`
		} `json:"Оценка"`
		Subtasks struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Подзадачи"`
		Deadline struct {
			Date struct {
				Start    string      `json:"start"`
				End      interface{} `json:"end"`
				TimeZone interface{} `json:"time_zone"`
			} `json:"date"`
		} `json:"Дедлайн"`
		Task struct {
			Title []struct {
				PlainText string `json:"plain_text"`
			} `json:"title"`
		} `json:"Task"`
	} `json:"properties"`
}

func (e *External) GetTasks(timeFilterType string, lastSynced int64, startCursor string, useTitleFilter bool) (tasks []entities.Task, lastUpdate int64, err error) {
	filter := buildFilter(timeFilterType, lastSynced, startCursor, useTitleFilter)

	lastUpdate = 0
	resp, err := notion.SearchPages(os.Getenv("TASKS_DB"), filter)
	if err != nil {
		return nil, 0, err
	}
	task := struct {
		Results    []Task `json:"results"`
		HasMore    bool   `json:"has_more"`
		NextCursor string `json:"next_cursor"`
	}{}

	json.Unmarshal(resp, &task)

	tasks = []entities.Task{}
	for _, w := range task.Results {
		tasks = append(tasks, entities.Task{
			ID: w.ID,
			Employee: func() string {
				if len(w.Properties.Worker.People) == 0 {
					return ""
				}
				return w.Properties.Worker.People[0].Name
			}(),
			Title: func() string {
				if len(w.Properties.Task.Title) == 0 {
					return ""
				}
				title := ""
				for _, t := range w.Properties.Task.Title {
					title += t.PlainText
				}

				return title
			}(),
			Status: w.Properties.Status.Status.Name,
			ProjectID: func() string {
				if len(w.Properties.Product.Relation) == 0 {
					return ""
				}
				return w.Properties.Product.Relation[0].ID
			}(),
			EmployeeID: func() string {
				if len(w.Properties.Worker.People) == 0 {
					return ""
				}
				return w.Properties.Worker.People[0].ID
			}(),
		})

		lastEditedTime, err := time.Parse(notion.TIME_LAYOUT_IN, w.LastEditedTime)
		if err != nil {
			return nil, 0, err
		}

		lastUpdate = lastEditedTime.Unix()
	}

	if task.HasMore {
		fmt.Println("has more")
		nextTasks, lastEditedTime, err := e.GetTasks(timeFilterType, lastSynced, task.NextCursor, useTitleFilter)
		if err != nil {
			return nil, 0, err
		}
		lastUpdate = lastEditedTime
		tasks = append(tasks, nextTasks...)
	}

	fmt.Println("Tasks: ", tasks)

	return tasks, lastUpdate, nil
}

func buildFilter(timeFilterType string, lastSynced int64, startCursor string, useTitleFilter bool) map[string]interface{} {
	filter := map[string]interface{}{
		"filter": map[string]interface{}{
			"timestamp": timeFilterType,
			timeFilterType: map[string]interface{}{
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

	if useTitleFilter {
		forbiddenWords := []string{
			"фикс", "пофиксить", "фиксить", "правка", "править", "поправить", "исправить", "правки", "исправление", "баг", "безуспешно", "разобраться",
		}

		titleFilter := []map[string]interface{}{}
		for _, word := range forbiddenWords {
			titleFilter = append(titleFilter, map[string]interface{}{
				"property": "Task",
				"rich_text": map[string]interface{}{
					"contains": word,
				},
			})
		}

		filter["filter"] = map[string]interface{}{
			"and": []map[string]interface{}{
				filter["filter"].(map[string]interface{}),
				{
					"or": titleFilter,
				},
			},
		}
	}

	if startCursor != "" {
		filter["start_cursor"] = startCursor
	}

	return filter
}

type Project struct {
	ID             string `json:"id"`
	CreatedTime    string `json:"created_time"`
	LastEditedTime string `json:"last_edited_time"`
	Icon           struct {
		Type     string   `json:"type"`
		External external `json:"external"`
		File     file     `json:"file"`
		Emoji    emoji    `json:"emoji"`
	} `json:"icon"`
	Properties struct {
		Name struct {
			Title []struct {
				PlainText string `json:"plain_text"`
			} `json:"title"`
		} `json:"Name"`
		Status struct {
			Status struct {
				Name string `json:"name"`
			} `json:"status"`
		} `json:"Статус"`
		Internal struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"internal"`
	} `json:"properties"`
}

type external struct {
	Url string `json:"url"`
}

type file struct {
	Url string `json:"url"`
}

type emoji struct {
	Emoji string `json:"emoji"`
}

func (e *External) GetProjects(lastSynced int64) (projects []entities.Project, lastUpdate int64, err error) {
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

	resp, err := notion.SearchPages(os.Getenv("PROJECTS_DB"), filter)
	if err != nil {
		return nil, 0, err
	}
	project := struct {
		Results []Project `json:"results"`
	}{}
	json.Unmarshal(resp, &project)

	projects = []entities.Project{}
	for _, w := range project.Results {
		projects = append(projects, entities.Project{
			ID: w.ID,
			Name: func() string {
				if len(w.Properties.Name.Title) == 0 {
					return ""
				}
				return w.Properties.Name.Title[0].PlainText
			}(),
			Icon: func() string {
				if w.Icon.Type == "emoji" {
					return w.Icon.Emoji.Emoji
				}
				if w.Icon.Type == "external" {
					return w.Icon.External.Url
				}
				if w.Icon.Type == "file" {
					return w.Icon.File.Url
				}
				return ""
			}(),
			IconType: w.Icon.Type,
			Status:   w.Properties.Status.Status.Name,
		})

		lastEditedTime, err := time.Parse(notion.TIME_LAYOUT_IN, w.LastEditedTime)
		if err != nil {
			return nil, 0, err
		}

		lastUpdate = lastEditedTime.Unix()
	}
	return projects, lastUpdate, nil
}

func (e *External) WriteOfTime(timeToWriteOf *entities.TimeMsg) error {

	req := map[string]interface{}{
		"Всего ч": map[string]interface{}{
			"number": float64(timeToWriteOf.Duration) / 60 / 60,
		},
		"Задача": map[string]interface{}{
			"relation": []map[string]interface{}{
				{
					"id": timeToWriteOf.TaskID,
				},
			},
		},
		"Что делали": map[string]interface{}{
			"type": "title",
			"title": []map[string]interface{}{
				{
					"type": "text",
					"text": map[string]interface{}{
						"content": timeToWriteOf.Description,
					},
				},
			},
		},
		"Дата работ": map[string]interface{}{
			"type": "date",
			"date": map[string]interface{}{
				"start": time.Now().Format(notion.TIME_LAYOUT),
			},
		},
		"Исполнитель": map[string]interface{}{
			"people": []map[string]interface{}{
				{
					"object": "user",
					"id":     timeToWriteOf.EmployeeID,
				},
			},
		},
	}

	_, err := notion.CreatePage(os.Getenv("TIMES_DB"), req, "")
	if err != nil {
		slog.Error("error creating time page in notion: " + err.Error())
		return err
	}
	return nil
}

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

func (e *External) GetTimes(timeFilterType string, lastSynced int64, startCursor string, useWhatDidFilter bool) (times []entities.Time, lastUpdate int64, err error) {
	filter := buildTimeFilter(timeFilterType, lastSynced, startCursor, useWhatDidFilter)

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

	if timeResults.HasMore {
		fmt.Println("time has more")
		nextTasks, lastEditedTime, err := e.GetTimes(timeFilterType, lastSynced, timeResults.NextCursor, useWhatDidFilter)
		if err != nil {
			return nil, 0, err
		}
		lastUpdate = lastEditedTime
		times = append(times, nextTasks...)
	}

	return times, lastUpdate, nil
}

func buildTimeFilter(timeFilterType string, lastSynced int64, startCursor string, useWhatDidFilter bool) map[string]interface{} {
	filter := map[string]interface{}{
		"filter": map[string]interface{}{
			"timestamp": timeFilterType,
			timeFilterType: map[string]interface{}{
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

	if useWhatDidFilter {
		forbiddenWords := []string{
			"фикс", "пофиксить", "фиксить", "правка", "править", "поправить", "исправить", "правки", "исправление", "баг", "безуспешно", "разобраться",
		}

		whatDidFilter := []map[string]interface{}{}
		for _, word := range forbiddenWords {
			whatDidFilter = append(whatDidFilter, map[string]interface{}{
				"property": "Что делали",
				"rich_text": map[string]interface{}{
					"contains": word,
				},
			})
		}

		filter["filter"] = map[string]interface{}{
			"and": []map[string]interface{}{
				filter["filter"].(map[string]interface{}),
				{
					"or": whatDidFilter,
				},
			},
		}
	}

	if startCursor != "" {
		filter["start_cursor"] = startCursor
	}

	return filter
}
