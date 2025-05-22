package storage

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
)

type Posts struct {
	log *slog.Logger
	DB  *sqlx.DB
}

func NewPostRepo(log *slog.Logger, db *sqlx.DB) *Posts {
	return &Posts{log: log, DB: db}
}
