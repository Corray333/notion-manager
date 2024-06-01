package main

import (
	"github.com/Corray333/notion-manager/internal/config"
	"github.com/Corray333/notion-manager/internal/server"
)

func main() {
	config.MustInit()
	// store := storage.NewStorage()

	// resp, _ := notion.GetPage("faf86e0c-c7e0-4bc7-93c5-037828d35c42")
	// var t notion.Task
	// json.Unmarshal(resp, &t)
	// fmt.Println(t)
	// if err := t.Update(store, &project.Project{
	// 	ProjectID:   "925e48a93ff54b4e99594805b5bfbfed",
	// 	Name:        "Экомобайл",
	// 	TimeDBID:    "4ca9a281ae6d49e7b859279809a30401",
	// 	TasksDBID:   "d98dbaea895f4fdebf3d2162d4db54f1",
	// 	WorkersDBID: "6eff59b93453498ca6087246c8ae186d",
	// }); err != nil {
	// 	fmt.Println(err)
	// }

	server.NewApp().Run()
}
