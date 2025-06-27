package service

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/arevbond/arevbond-blog/internal/service/blog/domain"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type PostRepository interface {
	All(ctx context.Context, limit int, offset int) ([]*domain.Post, error)
	Find(ctx context.Context, id int) (*domain.Post, error)
	Create(ctx context.Context, post *domain.Post) error
	Delete(ctx context.Context, id int) error
}

type Blog struct {
	log       *slog.Logger
	PostsRepo PostRepository
}

func New(log *slog.Logger, posts PostRepository) *Blog {
	return &Blog{log: log, PostsRepo: posts}
}

func (b *Blog) Posts(ctx context.Context, limit, offset int) ([]*domain.Post, error) {
	posts, err := b.PostsRepo.All(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("can't process posts in service: %w", err)
	}

	return posts, nil
}

func (b *Blog) Post(ctx context.Context, id int) (*domain.Post, error) {
	post, err := b.PostsRepo.Find(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("can't process post by id in service: %w", err)
	}

	return post, nil
}

func (b *Blog) CreatePost(ctx context.Context, params domain.PostParams) (int, error) {
	if params.Title == "" {
		params.Title = strings.TrimSuffix(params.Filename, filepath.Ext(params.Filename))
	}

	post := &domain.Post{
		ID:          0,
		Title:       params.Title,
		Description: params.Description,
		Content:     params.Content,
		Extension:   filepath.Ext(params.Filename),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := b.PostsRepo.Create(ctx, post)
	if err != nil {
		return -1, fmt.Errorf("can't create post: %w", err)
	}

	return post.ID, nil
}

func (b *Blog) MdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	//nolint:exhaustruct // default render options
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

func (b *Blog) DeletePost(ctx context.Context, id int) error {
	err := b.PostsRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}

	return nil
}
