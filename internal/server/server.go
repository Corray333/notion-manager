package server

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jomei/notionapi"
	"github.com/spf13/viper"
)

type App struct {
	NotionClient *notionapi.Client
}

func NewApp() *App {
	return &App{
		NotionClient: notionapi.NewClient(notionapi.Token(os.Getenv("NOTION_TOKEN"))),
	}
}

func (a *App) Run() {
	router := chi.NewMux()
	// store := storage.MustInit()

	slog.Info("Starting server on port " + viper.GetString("PORT"))
	if err := http.ListenAndServe(viper.GetString("PORT"), router); err != nil {
		panic(err)
	}
}
