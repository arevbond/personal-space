package service

import "log/slog"

type Auth struct {
	log        *slog.Logger
	adminToken string
}

func New(log *slog.Logger, adminToken string) *Auth {
	return &Auth{log: log, adminToken: adminToken}
}

func (a *Auth) IsAdminToken(token string) bool {
	return a.adminToken == token
}
