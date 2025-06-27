package server

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/arevbond/arevbond-blog/internal/middleware"
	"github.com/arevbond/arevbond-blog/internal/service/blog/domain"
)

type Blog interface {
	Posts(ctx context.Context, limit, offset int) ([]*domain.Post, error)
	Post(ctx context.Context, id int) (*domain.Post, error)
	CreatePost(ctx context.Context, params domain.PostParams) (int, error)
	DeletePost(ctx context.Context, id int) error
	ChangePublishStatus(ctx context.Context, id int, curPublishStatus bool) error

	MdToHTML(md []byte) []byte
}

func (s *Server) registerBlogRoutes(mux *http.ServeMux) {
	mux.Handle("GET /blog/posts", middleware.OptionalAuth(s.Auth, s.log)(http.HandlerFunc(s.postsPage)))
	mux.Handle("GET /blog/posts/{id}", middleware.OptionalAuth(s.Auth, s.log)(http.HandlerFunc(s.postPage)))

	mux.Handle("GET /blog/posts/form", middleware.RequireAuth(s.Auth, s.log)(http.HandlerFunc(s.createPostPage)))
	mux.Handle("POST /blog/posts", middleware.RequireAuth(s.Auth, s.log)(http.HandlerFunc(s.createPost)))
	mux.Handle("DELETE /blog/posts/{id}", middleware.RequireAuth(s.Auth, s.log)(http.HandlerFunc(s.deletePost)))
	mux.Handle("POST /blog/posts/{id}/toggle-publication",
		middleware.RequireAuth(s.Auth, s.log)(http.HandlerFunc(s.togglePostPublication)))
}

func (s *Server) postsPage(w http.ResponseWriter, r *http.Request) {
	isAdmin := r.Context().Value(middleware.IsAdminKey) != nil

	const pageLimit = 10

	posts, err := s.Blog.Posts(r.Context(), pageLimit, 0)
	if err != nil {
		s.log.Error("can't get posts from db", slog.Any("error", err))

		http.Error(w, "can't get posts from db", http.StatusInternalServerError)

		return
	}

	tmplData := struct {
		Posts   []*domain.Post
		IsAdmin bool
	}{
		Posts:   posts,
		IsAdmin: isAdmin,
	}

	s.renderTemplate(w, "posts.html", tmplData)
}

func (s *Server) postPage(w http.ResponseWriter, r *http.Request) {
	isAdmin := r.Context().Value(middleware.IsAdminKey) != nil

	idStr := r.PathValue("id")

	postID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)

		return
	}

	post, err := s.Blog.Post(r.Context(), postID)
	if err != nil {
		s.log.Error("can't process service post method", slog.Any("error", err))

		http.Error(w, "can't find post by id", http.StatusBadRequest)

		return
	}

	content := s.Blog.MdToHTML(post.Content)

	// #nosec G203 - Content is from trusted markdown stored in database
	tmplContent := template.HTML(content)

	tmplData := struct {
		ID          int
		Title       string
		Description string
		Content     template.HTML
		CreatedAt   string
		UpdatedAt   string
		IsPublished bool
		IsAdmin     bool
	}{
		ID:          post.ID,
		Title:       post.Title,
		Description: post.Description,
		Content:     tmplContent,
		CreatedAt:   post.CreatedAt.Format("02 Jan 2006"),
		UpdatedAt:   post.UpdatedAt.Format("02 Jan 2006"),
		IsPublished: post.IsPublished,
		IsAdmin:     isAdmin,
	}

	w.WriteHeader(http.StatusOK)

	s.renderTemplate(w, "post.html", tmplData)
}

func (s *Server) createPostPage(w http.ResponseWriter, r *http.Request) {
	s.renderTemplate(w, "create_post.html", nil)
}

func (s *Server) createPost(w http.ResponseWriter, r *http.Request) {
	// 1MB
	const maxRequestSize = 1_000_000

	if err := r.ParseMultipartForm(maxRequestSize); err != nil {
		s.log.Warn("can't parse file", slog.Any("error", err))

		http.Error(w, "can't parse form", http.StatusBadRequest)

		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")

	s.log.Debug("create post handler", slog.String("title", title),
		slog.String("description", description))

	file, header, err := r.FormFile("file")
	if err != nil {
		s.log.Warn("can't get file", slog.Any("error", err))

		http.Error(w, "can't get file from form", http.StatusBadRequest)

		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		s.log.Error("can't read file", slog.Any("error", err))

		http.Error(w, "can't read file", http.StatusInternalServerError)

		return
	}

	postParms := domain.PostParams{
		Title:       title,
		Description: description,
		Filename:    header.Filename,
		Content:     content,
		IsPublished: false,
	}

	postID, err := s.Blog.CreatePost(r.Context(), postParms)
	if err != nil {
		s.log.Error("can't create post", slog.Any("error", err))

		http.Error(w, "can't create post", http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/blog/posts/%d", postID), http.StatusFound)
}

func (s *Server) deletePost(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	postID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)

		return
	}

	err = s.Blog.DeletePost(r.Context(), postID)
	if err != nil {
		s.log.Error("delete post handler", slog.Any("error", err))

		http.Error(w, "server error", http.StatusInternalServerError)

		return
	}

	w.Header().Set("HX-Redirect", "/blog/posts")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) togglePostPublication(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	postID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)

		return
	}

	curPublishStatusStr := r.URL.Query().Get("is_published")

	curPublishStatus, err := strconv.ParseBool(curPublishStatusStr)
	if err != nil {
		s.log.Error("invalid publish status to toggle handler", slog.String("status", curPublishStatusStr),
			slog.Any("error", err))

		http.Error(w, "invalid status", http.StatusBadRequest)

		return
	}

	err = s.Blog.ChangePublishStatus(r.Context(), postID, curPublishStatus)
	if err != nil {
		s.log.Error("can't change publish status", slog.Any("error", err))

		http.Error(w, "server error", http.StatusInternalServerError)

		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/blog/posts/%d", postID))
	w.WriteHeader(http.StatusOK)
}
