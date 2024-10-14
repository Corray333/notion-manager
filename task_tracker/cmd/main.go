package main

import (
	"github.com/Corray333/task_tracker/internal/app"
	"github.com/Corray333/task_tracker/internal/config"
)

func main() {
	config.MustInit()

	app.New().Run()
}
