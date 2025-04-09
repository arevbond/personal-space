package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arevbond/arevbond-blog/internal/config"
	"github.com/arevbond/arevbond-blog/internal/server"
	storage "github.com/arevbond/arevbond-blog/internal/storaga"
)

// App contains all application dependency and launch http server.
type App struct {
	Server *server.Server
}

func New(log *slog.Logger, cfg config.Config) (*App, error) {
	db, err := storage.New(log, cfg.Storage)
	if err != nil {
		return nil, fmt.Errorf("can't create app: %w", err)
	}

	srv := server.New(log, cfg.Server, db).WithRoutes()

	return &App{
		Server: srv,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	if err := a.Server.Run(ctx); err != nil {
		return fmt.Errorf("app run: %w", err)
	}

	return nil
}
