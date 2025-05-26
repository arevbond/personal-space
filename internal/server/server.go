package server

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
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

//go:embed views/*
var templatesFS embed.FS

const (
	shutdownTimeout     = 5 * time.Second
	readerHeaderTimeout = 5 * time.Second
)

// Config содержит в себе зависимости для web сервера.
type Config struct {
}

type Server struct {
	*http.Server
	Config
	log  *slog.Logger
	tmpl *template.Template
}

func New(log *slog.Logger, cfg config.Server, dependency Config) *Server {
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	//nolint: exhaustruct // default options in http server is good
	srv := &http.Server{
		ReadHeaderTimeout: readerHeaderTimeout,
		Addr:              addr,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelError),
	}

	return &Server{
		Server: srv,
		Config: dependency,
		log:    log,
		tmpl:   template.Must(template.ParseFS(templatesFS, "views/*.html")),
	}
}

func (s *Server) WithRoutes() *Server {
	mux := http.NewServeMux()

	staticFS, err := fs.Sub(templatesFS, "views/static")
	if err != nil {
		s.log.Error("can't mount static directory", slog.Any("error", err))
	}

	if staticFS != nil {
		mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticFS)))
	}

	mux.HandleFunc("GET /ping", s.ping)

	mux.HandleFunc("GET /posts", s.htmlPosts)

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
