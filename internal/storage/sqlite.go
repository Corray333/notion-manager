package storage

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func MustInit() *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", "../notion.db")
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	return db
}
