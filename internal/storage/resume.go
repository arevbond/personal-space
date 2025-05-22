package storage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/arevbond/arevbond-blog/internal/models"
	"github.com/jmoiron/sqlx"
)

type Resume struct {
	DB  *sqlx.DB
	log *slog.Logger
}

func NewResumeRepo(log *slog.Logger, db *sqlx.DB) (*Resume, error) {
	return &Resume{
		DB:  db,
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

func (s *Resume) Resumes(ctx context.Context) ([]models.Resume, error) {
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

func (s *Resume) UploadResume(ctx context.Context, cv models.Resume) error {
	query := `INSERT INTO resumes (name, content, file_extension)
				VALUES ($1, $2, $3)`

	args := []any{cv.Name, cv.Content, cv.FileExtension}

	_, err := s.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("can't upload cv: %w", err)
	}

	return nil
}

func (s *Resume) RemoveResume(ctx context.Context, id int) error {
	query := `DELETE FROM resumes WHERE id = $1;`

	_, err := s.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("can't remove resume: %w", err)
	}

	return nil
}
