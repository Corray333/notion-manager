package server

import (
	"log/slog"
	"net/http"

	"github.com/Corray333/notion-manager/internal/server/handlers"
	"github.com/Corray333/notion-manager/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/jomei/notionapi"
	"github.com/spf13/viper"

	_ "github.com/Corray333/notion-manager/docs" // Import the generated docs package

	httpSwagger "github.com/swaggo/http-swagger"
)

type App struct {
	NotionClient *notionapi.Client
}

func NewApp() *App {
	return &App{}
}

func (a *App) Run() {
	router := chi.NewMux()
	store := storage.NewStorage()

	// router.Post("/projects", handlers.NewProject(store))
	router.Patch("/sync", handlers.UpdateDatabases(store))
	router.Get("/fix", handlers.GetToBeUpdated(store))

	// Swagger
	router.Get("/swagger/*", httpSwagger.WrapHandler)

	slog.Info("Starting server on port " + viper.GetString("PORT"))
	if err := http.ListenAndServe(viper.GetString("PORT"), router); err != nil {
		panic(err)
	}
}
