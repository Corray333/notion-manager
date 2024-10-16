package external

import (
	"encoding/json"
	"log"
	"log/slog"
	"os"
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
				if w.Icon.Type == "emoji" {
					return w.Icon.Emoji.Emoji
				} else if w.Icon.Type == "external" {
					return w.Icon.External.Url
				} else if w.Icon.Type == "file" {
					return w.Icon.File.Url
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

func (e *External) GetTasks(lastSynced int64, startCursor string) (tasks []entities.Task, lastUpdate int64, err error) {
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

	if startCursor != "" {
		filter["start_cursor"] = startCursor
	}

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
			ProjectID: func() *string {
				if len(w.Properties.Product.Relation) == 0 {
					return nil
				}
				return &w.Properties.Product.Relation[0].ID
			}(),
			EmployeeID: func() *string {
				if len(w.Properties.Worker.People) == 0 {
					return nil
				}
				return &w.Properties.Worker.People[0].ID
			}(),
		})

		lastEditedTime, err := time.Parse(notion.TIME_LAYOUT_IN, w.LastEditedTime)
		if err != nil {
			return nil, 0, err
		}

		lastUpdate = lastEditedTime.Unix()
	}

	// if task.HasMore {
	// 	nextTasks, err := e.GetTasks(lastSynced, task.NextCursor)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	tasks = append(tasks, nextTasks...)
	// }

	return tasks, lastUpdate, nil
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
	}

	_, err := notion.CreatePage(os.Getenv("TIME_DB"), req, "")
	if err != nil {
		slog.Error("error creating time page in notion: " + err.Error())
		return err
	}
	return nil
}
