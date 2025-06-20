package posts

import (
	"context"
	"fmt"
	"github.com/arevbond/arevbond-blog/internal/service/blog/domain"
	"github.com/arevbond/arevbond-blog/internal/service/blog/storage"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/arevbond/arevbond-blog/internal/config"
	"github.com/arevbond/arevbond-blog/internal/db"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type StorageSuite struct {
	suite.Suite
	log       *slog.Logger
	ctx       context.Context
	container *postgres.PostgresContainer
	conn      *sqlx.DB
	repo      *storage.Posts
}

func TestStorageSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip integration tests in short mode")
	}

	suite.Run(t, new(StorageSuite))
}

func (s *StorageSuite) SetupSuite() {
	s.log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	s.ctx = context.Background()

	cfg := config.Storage{
		DatabaseName: "test_db",
		User:         "test_user",
		Password:     "test_password",
	}

	container, err := postgres.Run(s.ctx, "postgres:17-alpine",
		postgres.WithDatabase(cfg.DatabaseName),
		postgres.WithUsername(cfg.User),
		postgres.WithPassword(cfg.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(15*time.Second)),
	)
	s.Require().NoError(err)
	s.container = container

	host, err := s.container.Host(s.ctx)
	s.Require().NoError(err)
	cfg.Host = host

	mappedPort, err := s.container.MappedPort(s.ctx, "5432/tcp")
	s.Require().NoError(err)
	port, err := strconv.Atoi(mappedPort.Port())
	s.Require().NoError(err)
	cfg.Port = port

	s.conn, err = db.NewConn(cfg)
	s.Require().NoError(err)
	s.Require().NoError(s.conn.Ping())

	err = migrate(s.conn)
	s.Require().NoError(err)

	s.repo = storage.NewPostsRepo(s.log, s.conn)
}

func (s *StorageSuite) TearDownSuite() {
	if s.conn != nil {
		s.Require().NoError(s.conn.Close())
	}
	if s.container != nil {
		s.Require().NoError(s.container.Terminate(s.ctx))
	}
}

func (s *StorageSuite) SetupTest() {
	s.truncateTables()
}

func (s *StorageSuite) truncateTables() {
	tables := []string{
		"posts",
	}

	for _, table := range tables {
		_, err := s.conn.ExecContext(s.ctx, fmt.Sprintf("TRUNCATE %s CASCADE;", table))
		s.Require().NoError(err)
	}
}

func migrate(conn *sqlx.DB) error {
	root := findRootDir()
	migrationPath := filepath.Join(root, "migrations")
	_, err := os.Stat(migrationPath)
	if err != nil {
		return fmt.Errorf("dir migrations doesn't exists: %w", err)
	}

	return goose.Up(conn.DB, migrationPath)
}

func findRootDir() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func (s *StorageSuite) TestCreate() {
	tests := []struct {
		name string
		post *domain.Post
	}{
		{
			name: "success creation post",
			post: &domain.Post{
				ID:          0,
				Title:       "title",
				Description: "desc",
				Content:     []byte("1110"),
				Extension:   ".md",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := s.repo.Create(s.ctx, tt.post)
			s.Require().NoError(err)

			s.Assert().NotEqual(0, tt.post.ID)
		})
	}

}
