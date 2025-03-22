package main

import (
	"flag"
	"log"
	"log/slog"

	"github.com/arevbond/arevbond-blog/internal/app"
	"github.com/arevbond/arevbond-blog/internal/config"
)

var configPath = flag.String("config", "configs/application.yaml", "path to application config")

func main() {
	flag.Parse()

	logger := slog.Default()

	cfg, err := config.New(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	logger.Info("server", slog.Any("server", cfg.Server))

	logger.Info("application started")

	app := app.New(logger, cfg)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
