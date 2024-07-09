package config

import (
	"log/slog"

	"github.com/Corray333/notion-manager/pkg/server/logger"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func MustInit() {
	if err := godotenv.Load("../.env"); err != nil {
		panic(err)
	}
	configPath := "../configs/dev.yml"
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	SetupLogger()
}

func SetupLogger() {
	handler := logger.NewHandler(nil)
	log := slog.New(handler)
	slog.SetDefault(log)
}
