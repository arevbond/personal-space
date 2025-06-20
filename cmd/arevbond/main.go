package main

import (
	"context"
	"flag"
	"io"
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

	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	logger := mustSetupLogger(cfg.Env)

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

func mustSetupLogger(env string) *slog.Logger {
	switch env {
	case config.EnvLocal:
		return setupPrettyLogger()
	case config.EnvProd:
		return setupJSONLogger(os.Stdout)
	default:
		log.Fatalf("can't setup logger: unknown environment: %s", env)
	}

	return nil
}

func setupJSONLogger(writers ...io.Writer) *slog.Logger {
	//nolint: exhaustruct // default slog handler
	logger := slog.New(slog.NewJSONHandler(io.MultiWriter(writers...), &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	return logger
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
