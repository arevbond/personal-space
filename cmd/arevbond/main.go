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

func main() {
	flag.Parse()

	logger := setupPrettyLogger()

	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	logger.Debug("http server", slog.Any("config", cfg.Server))

	logger.Info("application started", slog.String("Env", cfg.Env))

	application, err := app.New(logger, cfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	if err = application.Run(ctx); err != nil {
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
