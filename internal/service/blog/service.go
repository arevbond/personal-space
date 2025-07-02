package blog

import (
	"log/slog"

	"github.com/arevbond/arevbond-blog/internal/service/blog/service"
	"github.com/arevbond/arevbond-blog/internal/service/blog/service/processor"
	"github.com/arevbond/arevbond-blog/internal/service/blog/storage"
	"github.com/jmoiron/sqlx"
)

func NewBlogModule(log *slog.Logger, db *sqlx.DB) *service.Blog {
	postsRepo := storage.NewPostsRepo(log, db)
	imageProcessor := processor.NewImageProcessor(log)
	categoryRepo := storage.NewCategoriesRepo(log, db)

	return service.New(log, postsRepo, imageProcessor, categoryRepo)
}
