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
		SELECT id, title, description, content, extension, created_at, updated_at
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

func (p *Posts) Find(ctx context.Context, postID int) (*domain.Post, error) {
	query := `
		SELECT id, title, description, content, extension, created_at, updated_at
		FROM posts
		WHERE id = $1;`

	var post domain.Post

	err := p.DB.GetContext(ctx, &post, query, postID)
	if err != nil {
		return nil, fmt.Errorf("can't get post from db: %w", err)
	}

	return &post, nil
}

func (p *Posts) Create(ctx context.Context, post *domain.Post) error {
	query := `
		INSERT INTO posts (title, description, content, extension, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;`

	args := []any{post.Title, post.Description, post.Content, post.Extension, post.CreatedAt, post.UpdatedAt}

	row := p.DB.QueryRowContext(ctx, query, args...)
	if err := row.Scan(&post.ID); err != nil {
		return fmt.Errorf("can't scan id for post: %w", err)
	}

	if err := row.Err(); err != nil {
		return fmt.Errorf("row error in creation post: %w", err)
	}

	return nil
}
