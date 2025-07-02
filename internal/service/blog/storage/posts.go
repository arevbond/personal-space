package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/arevbond/arevbond-blog/internal/service/blog/domain"
	"github.com/arevbond/arevbond-blog/internal/service/errs"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

type Posts struct {
	log *slog.Logger
	DB  *sqlx.DB
}

const (
	UniqueViolationErr = "23505"
)

func NewPostsRepo(log *slog.Logger, db *sqlx.DB) *Posts {
	return &Posts{log: log, DB: db}
}

func (p *Posts) All(ctx context.Context, limit int, offset int, publishedOnly bool) ([]*domain.Post, error) {
	query := `
		SELECT p.id, title, description, content, extension, slug, is_published, category_id, 
		       c.name as category_name, created_at, updated_at
		FROM posts p
		LEFT JOIN categories c ON category_id = c.id
		WHERE ($3 = false OR is_published = true)
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2;`

	posts := []*domain.Post{}

	err := p.DB.SelectContext(ctx, &posts, query, limit, offset, publishedOnly)
	if err != nil {
		return nil, fmt.Errorf("can't get posts from db: %w", err)
	}

	return posts, nil
}

func (p *Posts) Find(ctx context.Context, postID int) (*domain.Post, error) {
	query := `
		SELECT p.id, title, description, content, extension, slug, is_published, category_id, 
		       c.name as category_name, created_at, updated_at
		FROM posts p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE id = $1;`

	var post domain.Post

	err := p.DB.GetContext(ctx, &post, query, postID)
	if err != nil {
		return nil, fmt.Errorf("can't get post from db: %w", err)
	}

	return &post, nil
}

func (p *Posts) FindBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	query := `
		SELECT p.id, title, description, content, extension, slug, is_published, category_id,
		       c.name as category_name, created_at, updated_at
		FROM posts p
		INNER JOIN categories c ON p.category_id = c.id
		WHERE slug = $1;`

	var post domain.Post

	err := p.DB.GetContext(ctx, &post, query, slug)
	if err != nil {
		return nil, fmt.Errorf("can't get post from db: %w", err)
	}

	return &post, nil
}

func (p *Posts) Create(ctx context.Context, post *domain.Post) error {
	query := `
		INSERT INTO posts (title, description, content, extension, slug, is_published, 
		                   category_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id;`

	args := []any{post.Title, post.Description, post.Content, post.Extension, post.Slug,
		post.IsPublished, post.CategoryID, post.CreatedAt, post.UpdatedAt}

	row := p.DB.QueryRowContext(ctx, query, args...)
	if err := row.Scan(&post.ID); err != nil {
		if IsErrorCode(err, UniqueViolationErr) {
			return fmt.Errorf("can't insert new row %w: %w", errs.ErrDuplicate, err)
		}

		return fmt.Errorf("can't scan id for post: %w", err)
	}

	if err := row.Err(); err != nil {
		return fmt.Errorf("row error in creation post: %w", err)
	}

	return nil
}

func (p *Posts) SetPublicationStatus(ctx context.Context, postID int, isPublished bool) error {
	query := `UPDATE posts
				SET is_published = $1
				WHERE id = $2;`

	args := []any{isPublished, postID}

	result, err := p.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("can't set publication status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("post with id %d: %w", postID, errs.ErrNotFound)
	}

	return nil
}

func (p *Posts) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM posts WHERE id = $1;`

	args := []any{id}

	_, err := p.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("can't delete post: %w", err)
	}

	return nil
}

func IsErrorCode(err error, errCode string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == errCode
	}

	return false
}

func (p *Posts) AllWithCategory(
	ctx context.Context,
	limit int,
	offset int,
	publishedOnly bool,
	categoryID int,
) ([]*domain.Post, error) {
	query := `
		SELECT p.id, title, description, content, extension, slug, is_published, category_id, 
		       c.name as category_name, created_at, updated_at
		FROM posts p
		LEFT JOIN categories c ON category_id = c.id
		WHERE ($3 = false OR is_published = true) AND p.category_id = $4
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2;`

	posts := []*domain.Post{}

	err := p.DB.SelectContext(ctx, &posts, query, limit, offset, publishedOnly, categoryID)
	if err != nil {
		return nil, fmt.Errorf("can't get posts from db: %w", err)
	}

	return posts, nil
}
