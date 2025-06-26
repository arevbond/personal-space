package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arevbond/arevbond-blog/internal/config"
	"github.com/arevbond/arevbond-blog/internal/db"
	"github.com/arevbond/arevbond-blog/internal/server"
	"github.com/arevbond/arevbond-blog/internal/service/auth"
	"github.com/arevbond/arevbond-blog/internal/service/blog"
)

// App contains all application dependency and launch http server.
type App struct {
	Server *server.Server
}

func New(log *slog.Logger, cfg config.Config) (*App, error) {
	conn, err := db.NewConn(cfg.Storage)
	if err != nil {
		return nil, fmt.Errorf("can't connect to storage: %w", err)
	}

	blogService := blog.NewBlogModule(log, conn)
	authService := auth.NewAuthModule(log, cfg.AdminToken, cfg.SecretKeyJWT)

	srv := server.New(log, cfg.Server, server.Services{Blog: blogService, Auth: authService})
	srv.ConfigureRoutes()

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
