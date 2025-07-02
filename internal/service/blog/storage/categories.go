package storage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arevbond/arevbond-blog/internal/service/blog/domain"
	"github.com/jmoiron/sqlx"
)

type Categories struct {
	log *slog.Logger
	DB  *sqlx.DB
}

func NewCategoriesRepo(log *slog.Logger, db *sqlx.DB) *Categories {
	return &Categories{log: log, DB: db}
}

func (c *Categories) All(ctx context.Context) ([]*domain.Category, error) {
	query := `SELECT id, name 
			  FROM categories;`

	categories := []*domain.Category{}

	err := c.DB.SelectContext(ctx, &categories, query)
	if err != nil {
		return nil, fmt.Errorf("can't get categories: %w", err)
	}

	return categories, nil
}
