package main

import (
	"fmt"
	"time"

	"github.com/Corray333/notion-manager/internal/config"
	"github.com/Corray333/notion-manager/internal/gsheets"
	"github.com/Corray333/notion-manager/internal/server"
	"github.com/robfig/cron/v3"
)

func main() {

	config.MustInit()

	// projectName, tasks, err := mindmap.ParseMarkdownTasks("Хронодокс.md")
	// if err != nil {
	// 	panic(err)
	// }

	// notion.CreateMindmapTasks(projectName, tasks)

	// go gsheets.UpdateGoogleSheets()

	// file, _ := os.ReadFile("test.md")
	// _, tasks, _ := mindmap.ParseMarkdownTasks(string(file))
	// j, _ := json.MarshalIndent(tasks, "", "    ")
	// fmt.Println(string(j))

	c := cron.New(cron.WithLocation(time.FixedZone("MSK", 3*60*60)))

	_, err := c.AddFunc("0 5 * * *", func() { gsheets.UpdateGoogleSheets() })
	if err != nil {
		fmt.Println("Error scheduling function - ", err)
		return
	}

	c.Start()

	server.NewApp().Run()

	// fmt.Println(gsheets.UpdateGoogleSheets())

}
