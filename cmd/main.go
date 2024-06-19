package main

import (
	"github.com/Corray333/notion-manager/internal/config"
	"github.com/Corray333/notion-manager/internal/server"
)

func main() {

	config.MustInit()

	server.NewApp().Run()
}
