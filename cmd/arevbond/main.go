package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/arevbond/arevbond-blog/internal/app"
	"github.com/arevbond/arevbond-blog/internal/config"
)

var configPath = flag.String("config", "configs/application.yaml", "path to application config")

func main() {
	flag.Parse()

	logger := setupLogger()

	cfg, err := config.New(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	logger.Debug("http server", slog.Any("config", cfg.Server))

	logger.Info("application started")

	app := app.New(logger, cfg)

	ctx := context.Background()

	if err := app.Run(ctx); err != nil {
		panic(err)
	}
}

func setupLogger() *slog.Logger {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   false,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))

	return logger
}
