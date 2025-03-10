package main

import "log/slog"

func main() {
	logger := slog.Default()

	logger.Info("application started")
}
