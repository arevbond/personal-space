package server

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/arevbond/arevbond-blog/internal/config"
	"github.com/arevbond/arevbond-blog/internal/models"
)

//go:embed views/*
var templatesFS embed.FS

const (
	shutdownTimeout     = 5 * time.Second
	readerHeaderTimeout = 5 * time.Second
)

type CVManager interface {
	ListCV(ctx context.Context) ([]models.CV, error)
	UploadCV(ctx context.Context, cv models.CV) error
}

type Server struct {
	*http.Server
	log     *slog.Logger
	tmpl    *template.Template
	manager CVManager
}

func New(log *slog.Logger, cfg config.Server, manager CVManager) *Server {
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	//nolint: exhaustruct // default options in http server is good
	srv := &http.Server{
		ReadHeaderTimeout: readerHeaderTimeout,
		Addr:              addr,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelError),
	}

	return &Server{
		Server:  srv,
		log:     log,
		tmpl:    template.Must(template.ParseFS(templatesFS, "views/*.html")),
		manager: manager,
	}
}

func (s *Server) WithRoutes() *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", s.ping)
	// mux.HandleFunc("GET /cv", s.htmlCVpreview)
	mux.HandleFunc("GET /cv", s.htmlAllCV)
	mux.HandleFunc("POST /cv", s.uploadcv)
	mux.HandleFunc("GET /", s.htmlIndex)

	s.Handler = mux

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
