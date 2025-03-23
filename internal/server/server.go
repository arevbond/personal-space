package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/arevbond/arevbond-blog/internal/config"
)

const (
	shutdownTimeout  = 5 * time.Second
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

func (s *Server) Run(ctx context.Context) error {
	idleConnsClosed := make(chan struct{})

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-ctx.Done():
		case <-signals:
		}

		if err := s.Shutdown(ctx); err != nil {
			s.log.Error("graceful shutdown http server", slog.Any("error", err))
		}

		close(idleConnsClosed)
	}()

	if err := s.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			s.log.Info("http server stopped")
		} else {
			return fmt.Errorf("server run: %w", err)
		}
	}

	<-idleConnsClosed

	return nil
}
