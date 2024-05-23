package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/Corray333/notion-manager/internal/config"
	"github.com/Corray333/notion-manager/internal/notion"
	"github.com/Corray333/notion-manager/internal/project"
	"github.com/Corray333/notion-manager/internal/storage"
)

type TableType string

const (
	TimeTable TableType = "time"
	TaskTable TableType = "task"
)

type Error struct {
	err        error
	project    project.Project
	table_type TableType
	id         string
}

func (e Error) String() string {
	return fmt.Sprintf("Error: %s, Table: %s, Project: %s, ID: %s", e.err.Error(), e.table_type, e.project.Name, e.id)
}

func main() {
	config.MustInit()

	store := storage.NewStorage()

	projects, err := store.GetProjects()
	if err != nil {
		panic(err)
	}

	errs := []Error{}

	for _, project := range projects {
		project.TasksLastSynced = 1668038400
		tasks, err := notion.GetTasks(store, project, "")
		if err != nil {
			errs = append(errs, Error{
				err:        errors.Join(errors.New("error while getting tasks: "), err),
				table_type: TaskTable,
				project:    project,
			})
		}
		fmt.Printf("Loaded %d tasks.\n", len(tasks))
		for _, task := range tasks {
			err := task.Upload(store, &project)
			if err != nil {
				fmt.Println(err.Error())
				errs = append(errs, Error{
					err:        err,
					table_type: TaskTable,
					project:    project,
					id:         task.ID,
				})
			}
		}

		times, err := notion.GetTimes(store, project)
		if err != nil {
			errs = append(errs, Error{
				err:        errors.Join(errors.New("error while getting time rows: "), err),
				table_type: TimeTable,
				project:    project,
			})
		}
		fmt.Printf("Loaded %d times.", len(times))
		for _, time := range times {
			if err := time.Upload(store, project); err != nil {
				fmt.Println(err.Error())
				errs = append(errs, Error{
					err:        err,
					table_type: TaskTable,
					project:    project,
					id:         time.ID,
				})
			}
		}

		store.SetLastSynced(project)
		logs, _ := os.Create("logs.txt")
		for _, err := range errs {
			logs.WriteString(err.String() + "\n")
		}
		logs.Close()
	}

	//
	//
	//
	//
	//
	//

	// resp, _ := notion.GetWorker("6eff59b93453498ca6087246c8ae186d", "c767b21a-f61b-4edf-b71c-ae99ec9fd51f")
	// println(resp.ID)

	//
	//
	//
	//
	//
	//

	// req := map[string]interface{}{
	// 	"Name": map[string]interface{}{
	// 		"type": "title",
	// 		"title": []map[string]interface{}{
	// 			{
	// 				"type": "text",
	// 				"text": map[string]interface{}{
	// 					"content": "Test",
	// 				},
	// 			},
	// 		},
	// 	},
	// 	"Всего ч": map[string]interface{}{
	// 		"type":   "number",
	// 		"number": 10,
	// 	},
	// 	"Задача": map[string]interface{}{
	// 		"type": "relation",
	// 		"relation": []map[string]interface{}{
	// 			{
	// 				"id": "621930b9-aa50-4817-85a0-e55eab4f0c47",
	// 			},
	// 		},
	// 	},
	// }
	// resp, err := notion.CreatePage("4ca9a281ae6d49e7b859279809a30401", req)
	// if err != nil {
	// 	panic(err)
	// }
	// println(string(resp))

	// store := storage.NewStorage()
	// projects, err := store.GetProjects()
	// if err != nil {
	// 	panic(err)
	// }

	// for _, project := range projects {

	// }

	// req := map[string]interface{}{
	// 	"filter": map[string]interface{}{
	// 		"timestamp": "created_time",
	// 		"created_time": map[string]interface{}{
	// 			"on_or_after": "2024-05-05",
	// 		},
	// 	},
	// }

	// resp, err := notion.SearchPages("4ca9a281ae6d49e7b859279809a30401", req)
	// if err != nil {
	// 	panic(err)

	// }

	// fmt.Println(string(resp))

	// worker, _ := notion.GetWorker("6eff59b93453498ca6087246c8ae186d", "05186f0f-68e4-4073-b048-4089ebbd3381")
	// fmt.Println(worker.ID)

	// resp, err := notion.GetPage("268c487139fd4c7896814d62ae34dcee")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(resp))

	// data := map[string]interface{}{}
	// if err := json.Unmarshal(resp, &data); err != nil {
	// 	panic(err)
	// }
	// fmt.Println()
	// fmt.Println("Test: ", data)
	// fmt.Println()

	// resp, err = notion.CreatePage("4ca9a281ae6d49e7b859279809a30401", map[string]interface{}{
	// 	"Всего ч": data["results"].([]interface{})[0].(map[string]interface{})["properties"].(map[string]interface{})["Всего ч"],
	// 	"Задача":  data["results"].([]interface{})[0].(map[string]interface{})["properties"].(map[string]interface{})["Задача"],
	// 	"Name":    data["results"].([]interface{})[0].(map[string]interface{})["properties"].(map[string]interface{})["Name"],
	// })
	// if err != nil {
	// 	panic(err)
	// }
	// println(string(resp))

	// config.MustInit()
	// server.NewApp().Run()
}
