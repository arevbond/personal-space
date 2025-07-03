package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/arevbond/arevbond-blog/internal/service/blog/domain"
	"github.com/arevbond/arevbond-blog/internal/service/errs"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type PostRepository interface {
	All(ctx context.Context, limit int, offset int, publishedOnly bool) ([]*domain.Post, error)
	AllWithCategory(ctx context.Context, limit int, offset int, publishedOnly bool, categoryID int) ([]*domain.Post, error)
	Find(ctx context.Context, id int) (*domain.Post, error)
	FindBySlug(ctx context.Context, slug string) (*domain.Post, error)
	Create(ctx context.Context, post *domain.Post) error
	Update(ctx context.Context, params domain.UpdatePostParams) error
	Delete(ctx context.Context, id int) error

	SetPublicationStatus(ctx context.Context, id int, isPublished bool) error
}

type CategoriesRepository interface {
	All(ctx context.Context) ([]*domain.Category, error)
}

type ImageProcessor interface {
	AddPrefix(content []byte, prefix string) ([]byte, error)
}

type Blog struct {
	log            *slog.Logger
	PostsRepo      PostRepository
	CategoriesRepo CategoriesRepository
	ImageProcessor ImageProcessor
}

func New(log *slog.Logger, posts PostRepository, imgReplacer ImageProcessor, categoryRepo CategoriesRepository) *Blog {
	return &Blog{log: log, PostsRepo: posts, ImageProcessor: imgReplacer, CategoriesRepo: categoryRepo}
}

func (b *Blog) Posts(ctx context.Context, params domain.SelectPostsParams) ([]*domain.Post, error) {
	publishedOnly := !params.IsAdmin

	var posts []*domain.Post

	var err error

	if params.CategoryID == 0 {
		posts, err = b.PostsRepo.All(ctx, params.Limit, params.Offset, publishedOnly)
		if err != nil {
			return nil, fmt.Errorf("can't process all posts in service: %w", err)
		}
	} else {
		posts, err = b.PostsRepo.AllWithCategory(ctx, params.Limit, params.Offset, publishedOnly, params.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("can't process all posts in service: %w", err)
		}
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

func (b *Blog) PostBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	post, err := b.PostsRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("can't process post by slug in service: %w", err)
	}

	return post, nil
}

func (b *Blog) CreatePost(ctx context.Context, params domain.CreatePostParams) (*domain.Post, error) {
	if params.Title == "" {
		params.Title = strings.TrimSuffix(params.Filename, filepath.Ext(params.Filename))
	}

	contentWithCorrectImages, err := b.ImageProcessor.AddPrefix(params.Content, "/static/images/")
	if err != nil {
		return nil, fmt.Errorf("can't add prefix to image: %w", err)
	}

	var lastError error

	baseSlug := params.Title
	if params.Slug != "" {
		baseSlug = params.Slug
	}

	for i := 1; i <= 100; i++ {
		post := &domain.Post{
			ID:           0,
			Title:        params.Title,
			Description:  params.Description,
			Content:      contentWithCorrectImages,
			Extension:    filepath.Ext(params.Filename),
			IsPublished:  params.IsPublished,
			CategoryID:   params.CategoryID,
			CategoryName: "", // не используется при создании нового поста
			Slug:         b.covertTitleToSlug(baseSlug, i),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		lastError = b.PostsRepo.Create(ctx, post)
		if nil == lastError {
			return post, nil
		}

		if !errors.Is(lastError, errs.ErrDuplicate) {
			return nil, fmt.Errorf("can't create post: %w", lastError)
		}
	}

	return nil, fmt.Errorf("can't create post: %w", lastError)
}

func (b *Blog) covertTitleToSlug(title string, num int) string {
	lowerTitle := strings.ToLower(title)
	strs := strings.Fields(b.removeSpecialChars(lowerTitle))

	enStrs := make([]string, len(strs))

	for i := 0; i < len(enStrs); i++ {
		enStrs[i] = b.ruToEn(strs[i])
	}

	slug := strings.Join(enStrs, "-")

	if num > 1 {
		slug = fmt.Sprintf("%s-%d", slug, num)
	}

	return slug
}

func (b *Blog) ruToEn(str string) string {
	rutoEnMp := map[string]string{
		"а": "a",
		"б": "b",
		"в": "v",
		"г": "g",
		"д": "d",
		"е": "e",
		"ж": "zh",
		"з": "z",
		"и": "i",
		"й": "i",
		"к": "k",
		"л": "l",
		"м": "m",
		"н": "n",
		"о": "o",
		"п": "p",
		"р": "r",
		"с": "s",
		"т": "t",
		"у": "u",
		"ф": "f",
		"х": "kh",
		"ц": "ts",
		"ч": "ch",
		"ш": "sh",
		"щ": "shch",
		"ъ": "",
		"ы": "y",
		"ь": "",
		"э": "e",
		"ю": "iu",
		"я": "ia",
		"ё": "e",
	}

	var sb strings.Builder

	for _, ch := range str {
		if res, ok := rutoEnMp[string(ch)]; ok {
			sb.WriteString(res)
		} else {
			sb.WriteRune(ch)
		}
	}

	return sb.String()
}

func (b *Blog) removeSpecialChars(str string) string {
	specialChars := map[rune]struct{}{'.': {}, ',': {}, '|': {}, '!': {}, '?': {}, ':': {}, ';': {}, '"': {}, '\'': {}}

	var sb strings.Builder

	for _, ch := range str {
		if _, ok := specialChars[ch]; ok {
			continue
		}

		sb.WriteRune(ch)
	}

	return sb.String()
}

func (b *Blog) UpdatePost(ctx context.Context, params domain.UpdatePostParams) error {
	contentWithCorrectImages, err := b.ImageProcessor.AddPrefix(params.Content, "/static/images/")
	if err != nil {
		return fmt.Errorf("can't add prefix to image: %w", err)
	}

	params.Content = contentWithCorrectImages

	err = b.PostsRepo.Update(ctx, params)
	if err != nil {
		return fmt.Errorf("service: %w", err)
	}

	return nil
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

func (b *Blog) ChangePublishStatus(ctx context.Context, id int, curPublishStatus bool) error {
	err := b.PostsRepo.SetPublicationStatus(ctx, id, !curPublishStatus)
	if err != nil {
		return fmt.Errorf("repository error: %w", err)
	}

	return nil
}

func (b *Blog) Categories(ctx context.Context) ([]*domain.Category, error) {
	categories, err := b.CategoriesRepo.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("blog: %w", err)
	}

	return categories, nil
}
