package config

import (
	"log/slog"

	"github.com/Corray333/task_tracker/pkg/logger"
	"github.com/joho/godotenv"
)

func MustInit() {
	if err := godotenv.Load("../.env"); err != nil {
		panic("error while loading .env file: " + err.Error())
	}
	SetupLogger()
}

func SetupLogger() {
	handler := logger.NewHandler(nil)
	log := slog.New(handler)
	slog.SetDefault(log)
}
