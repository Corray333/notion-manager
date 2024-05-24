package server

import (
	"log/slog"
	"net/http"

	"github.com/Corray333/notion-manager/internal/server/handlers"
	"github.com/Corray333/notion-manager/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/jomei/notionapi"
	"github.com/spf13/viper"
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

	router.Post("/projects", handlers.NewProject(store))

	slog.Info("Starting server on port " + viper.GetString("PORT"))
	if err := http.ListenAndServe(viper.GetString("PORT"), router); err != nil {
		panic(err)
	}
}
