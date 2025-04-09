package service

import "log/slog"

type CVManager interface {
}

type CV struct {
	log     *slog.Logger
	manager CVManager
}
