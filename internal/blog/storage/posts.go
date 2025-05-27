package storage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arevbond/arevbond-blog/internal/blog/domain"
	"github.com/jmoiron/sqlx"
)

type Posts struct {
	log *slog.Logger
	DB  *sqlx.DB
}

func NewPostsRepo(log *slog.Logger, db *sqlx.DB) *Posts {
	return &Posts{log: log, DB: db}
}

func (p *Posts) All(ctx context.Context, limit int, offset int) ([]*domain.Post, error) {
	query := `
		SELECT id, title, description, body, created_at, updated_at
		FROM posts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2;`

	posts := []*domain.Post{}

	err := p.DB.SelectContext(ctx, &posts, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("can't get posts from db: %w", err)
	}

	return posts, nil
}
