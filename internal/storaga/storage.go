package storage

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"time"

	"github.com/arevbond/arevbond-blog/internal/config"
	"github.com/arevbond/arevbond-blog/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	DB  *sqlx.DB
	log *slog.Logger
}

func New(log *slog.Logger, cfg config.Storage) (*Storage, error) {
	hostWithPort := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	uri := fmt.Sprintf("postgresql://%s:%s@%s/%s", cfg.User, cfg.Password,
		hostWithPort, cfg.DatabaseName)
	connStr, err := pgx.ParseConfig(uri)

	if err != nil {
		return nil, fmt.Errorf("can't parse pg uri: %w", err)
	}

	pgxdb := stdlib.OpenDB(*connStr)

	if err = pgxdb.Ping(); err != nil {
		return nil, fmt.Errorf("can't ping db: %w", err)
	}

	return &Storage{
		DB:  sqlx.NewDb(pgxdb, "pgx"),
		log: log,
	}, nil
}

type CVEntity struct {
	ID            int       `db:"id"`
	Name          string    `db:"name"`
	Content       []byte    `db:"content"`
	FileExtension string    `db:"file_extension"`
	UpdatedAt     time.Time `db:"last_updated_at"`
}

func (c CVEntity) toModel() models.Resume {
	return models.Resume{
		ID:            c.ID,
		Name:          c.Name,
		Content:       c.Content,
		FileExtension: c.FileExtension,
		UpdatedAt:     c.UpdatedAt,
	}
}

func (s *Storage) Resumes(ctx context.Context) ([]models.Resume, error) {
	query := `SELECT id, name, content, file_extension, last_updated_at 
				FROM resumes
				ORDER BY last_updated_at DESC;`

	var entities []CVEntity

	err := s.DB.SelectContext(ctx, &entities, query)
	if err != nil {
		return nil, fmt.Errorf("can't select all cv: %w", err)
	}

	result := make([]models.Resume, 0, len(entities))

	for _, entity := range entities {
		result = append(result, entity.toModel())
	}

	return result, nil
}

func (s *Storage) UploadResume(ctx context.Context, cv models.Resume) error {
	query := `INSERT INTO resumes (name, content, file_extension)
				VALUES ($1, $2, $3)`

	args := []any{cv.Name, cv.Content, cv.FileExtension}

	_, err := s.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("can't upload cv: %w", err)
	}

	return nil
}

func (s *Storage) RemoveResume(ctx context.Context, id int) error {
	query := `DELETE FROM resumes WHERE id = $1;`

	_, err := s.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("can't remove resume: %w", err)
	}

	return nil
}
