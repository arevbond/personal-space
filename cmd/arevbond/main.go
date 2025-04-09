package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/arevbond/arevbond-blog/internal/app"
	"github.com/arevbond/arevbond-blog/internal/config"
	"github.com/lmittmann/tint"
)

var configPath = flag.String("config", "configs/application.yaml", "path to application config")

func main() {
	flag.Parse()

	logger := setupPrettyLogger()

	cfg, err := config.New(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	logger.Debug("http server", slog.Any("config", cfg.Server))

	logger.Info("application started")

	app, err := app.New(logger, cfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	if err := app.Run(ctx); err != nil {
		panic(err)
	}
}

func setupPrettyLogger() *slog.Logger {
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		AddSource:   false,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
		TimeFormat:  time.Kitchen,
		NoColor:     false,
	}))

	return logger
}

// func setupLogger() *slog.Logger {
// 	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
// 		AddSource:   false,
// 		Level:       slog.LevelDebug,
// 		ReplaceAttr: nil,
// 	}))

// 	return logger
// }
