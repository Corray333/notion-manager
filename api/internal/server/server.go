package server

import (
	"log/slog"
	"net/http"

	"github.com/Corray333/notion-manager/internal/server/handlers"
	"github.com/Corray333/notion-manager/internal/storage"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
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
	router.Use(middleware.Logger)

	// TODO: get allowed origins, headers and methods from cfg
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://*", "https://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Set-Cookie", "Refresh", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300, // Максимальное время кеширования предзапроса (в секундах)
	}))

	store := storage.NewStorage()

	// router.Post("/projects", handlers.NewProject(store))
	router.Patch("/api/sync", handlers.UpdateDatabases(store))
	router.Patch("/api/sheets", handlers.UpdateGoogleSheets)
	router.Get("/api/fix", handlers.GetToBeUpdated(store))
	router.Post("/api/mindmap", handlers.ParseMindmap)

	// Swagger
	router.Get("/api/swagger/*", httpSwagger.WrapHandler)

	slog.Info("Starting server on port " + viper.GetString("PORT"))
	if err := http.ListenAndServe(viper.GetString("PORT"), router); err != nil {
		panic(err)
	}
}
