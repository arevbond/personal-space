package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arevbond/arevbond-blog/internal/config"
	"github.com/arevbond/arevbond-blog/internal/server"
	storage "github.com/arevbond/arevbond-blog/internal/storage"
)

// App contains all application dependency and launch http server.
type App struct {
	Server *server.Server
}

func New(log *slog.Logger, cfg config.Config) (*App, error) {
	_, err := storage.NewConn(cfg.Storage)
	if err != nil {
		return nil, fmt.Errorf("can't connect to storage: %w", err)
	}

	srv := server.New(log, cfg.Server, server.Config{}).WithRoutes()

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
