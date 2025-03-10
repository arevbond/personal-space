package server

import (
	"log/slog"
	"net/http"
)

type Server struct {
	*http.Server
	log *slog.Logger
}

func New(log *slog.Logger) *Server {
	return &Server{log: log}
}

func (s *Server) ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("pong"))
}
