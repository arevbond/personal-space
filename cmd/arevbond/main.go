package main

import (
	"context"
	"flag"
	"log"
	"log/slog"

	"github.com/arevbond/arevbond-blog/internal/app"
	"github.com/arevbond/arevbond-blog/internal/config"
)

var configPath = flag.String("config", "configs/application.yaml", "path to application config")

func main() {
	flag.Parse()

	// TODO: make custom logger
	logger := slog.Default()

	cfg, err := config.New(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	logger.Info("server", slog.Any("server", cfg.Server))

	logger.Info("application started")

	app := app.New(logger, cfg)

	ctx := context.Background()

	if err := app.Run(ctx); err != nil {
		panic(err)
	}
}
