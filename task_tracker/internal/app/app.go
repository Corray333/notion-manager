package app

import (
	_ "github.com/Corray333/task_tracker/docs"
	"github.com/Corray333/task_tracker/internal/external"
	"github.com/Corray333/task_tracker/internal/repository"
	"github.com/Corray333/task_tracker/internal/service"
	"github.com/Corray333/task_tracker/internal/transport"
)

type app struct {
	store     *repository.Storage
	service   *service.Service
	transport *transport.Transport
}

func New() *app {

	storage := repository.New()
	external := external.New()
	service := service.New(storage, external)

	transport := transport.New(service)
	transport.RegisterRoutes()

	app := &app{
		store:     storage,
		service:   service,
		transport: transport,
	}

	return app
}

func (app *app) Run() {
	go app.service.Run()
	app.transport.Run()
}
