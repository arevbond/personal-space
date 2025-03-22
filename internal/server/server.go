package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/arevbond/arevbond-blog/internal/config"
)

const (
	readHeaderTimeot = 5 * time.Second
)

type Server struct {
	*http.Server
	log *slog.Logger
}

func New(log *slog.Logger, cfg config.Server) *Server {
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	//nolint: exhaustruct // default options in http server is good
	srv := &http.Server{
		ReadHeaderTimeout: readHeaderTimeot,
		Addr:              addr,
	}

	return &Server{
		Server: srv,
		log:    log,
	}
}

func (s *Server) WithRoutes() *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", s.ping)

	s.Server.Handler = mux

	return s
}

func (s *Server) Run() error {
	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server run: %w", err)
	}

	return nil
}
