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

	"github.com/arevbond/arevbond-blog/internal/blog/domain"
	"github.com/arevbond/arevbond-blog/internal/config"
)

//go:embed views/*
var templatesFS embed.FS

const (
	shutdownTimeout     = 5 * time.Second
	readerHeaderTimeout = 5 * time.Second
)

type Blog interface {
	Posts(ctx context.Context, limit, offset int) ([]*domain.Post, error)
}

// Services содержит в себе зависимости для web сервера.
type Services struct {
	Blog Blog
}

type Server struct {
	*http.Server
	Services
	log  *slog.Logger
	tmpl *template.Template
}

func New(log *slog.Logger, cfg config.Server, dependency Services) *Server {
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	//nolint: exhaustruct // default options in http server is good
	srv := &http.Server{
		ReadHeaderTimeout: readerHeaderTimeout,
		Addr:              addr,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelError),
	}

	return &Server{
		Server:   srv,
		Services: dependency,
		log:      log,
		tmpl: template.Must(template.ParseFS(templatesFS,
			"views/*.html", "views/blog/*.html")),
	}
}

func (s *Server) ConfigureRoutes() {
	mux := http.NewServeMux()

	staticFS, err := fs.Sub(templatesFS, "views/static")
	if err != nil {
		s.log.Error("can't mount static directory", slog.Any("error", err))
	}

	if staticFS != nil {
		mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticFS)))
	}

	mux.HandleFunc("GET /ping", s.ping)
	mux.HandleFunc("GET /", s.htmlIndex)

	s.registerBlogRoutes(mux)

	s.Handler = mux
}

func (s *Server) Run(ctx context.Context) error {
	idleConnClosed := make(chan struct{})

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

		close(idleConnClosed)
	}()

	if err := s.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			s.log.Info("http server stopped")
		} else {
			return fmt.Errorf("server run: %w", err)
		}
	}

	<-idleConnClosed

	return nil
}
