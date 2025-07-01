package server

import (
	"context"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/arevbond/arevbond-blog/internal/middleware"
	"github.com/arevbond/arevbond-blog/internal/service/blog/domain"
)

type Blog interface {
	Posts(ctx context.Context, limit, offset int, isAdmin bool) ([]*domain.Post, error)
	Post(ctx context.Context, id int) (*domain.Post, error)
	PostBySlug(ctx context.Context, slug string) (*domain.Post, error)
	CreatePost(ctx context.Context, params domain.PostParams) (*domain.Post, error)
	DeletePost(ctx context.Context, id int) error
	ChangePublishStatus(ctx context.Context, id int, curPublishStatus bool) error

	MdToHTML(md []byte) []byte
}

func (s *Server) registerBlogRoutes(mux *http.ServeMux) {
	mux.Handle("GET /blog/posts", middleware.OptionalAuth(s.Auth, s.log)(http.HandlerFunc(s.postsPage)))
	mux.Handle("GET /blog/posts/more", middleware.OptionalAuth(s.Auth, s.log)(http.HandlerFunc(s.posts)))
	mux.Handle("GET /blog/posts/{slug}", middleware.OptionalAuth(s.Auth, s.log)(http.HandlerFunc(s.postPage)))

	mux.Handle("GET /blog/posts/form", middleware.RequireAuth(s.Auth, s.log)(http.HandlerFunc(s.createPostPage)))
	mux.Handle("POST /blog/posts", middleware.RequireAuth(s.Auth, s.log)(http.HandlerFunc(s.createPost)))
	mux.Handle("DELETE /blog/posts/{id}", middleware.RequireAuth(s.Auth, s.log)(http.HandlerFunc(s.deletePost)))
	mux.Handle("POST /blog/posts/{id}/toggle-publication",
		middleware.RequireAuth(s.Auth, s.log)(http.HandlerFunc(s.togglePostPublication)))
}

func (s *Server) postsPage(w http.ResponseWriter, r *http.Request) {
	isAdmin := r.Context().Value(middleware.IsAdminKey) != nil

	posts, err := s.Blog.Posts(r.Context(), s.pageLimit+1, 0, isAdmin)
	if err != nil {
		s.log.Error("can't get posts from db", slog.Any("error", err))

		http.Error(w, "can't get posts from db", http.StatusInternalServerError)

		return
	}

	tmplData := PostsPageData{
		Posts:        posts,
		IsAdmin:      isAdmin,
		HasNextPages: false,
		NextOffset:   len(posts),
	}

	if len(posts) == s.pageLimit+1 {
		tmplData.HasNextPages = true
		tmplData.Posts = tmplData.Posts[:len(tmplData.Posts)-1]
		tmplData.NextOffset = len(tmplData.Posts)
	}

	s.renderTemplate(w, "posts.html", tmplData)
}

func (s *Server) posts(w http.ResponseWriter, r *http.Request) {
	isAdmin := r.Context().Value(middleware.IsAdminKey) != nil

	offsetStr := r.URL.Query().Get("offset")

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		s.log.Error("can't convert offset to int", slog.Any("error", err))

		offset = 0
	}

	posts, err := s.Blog.Posts(r.Context(), s.pageLimit+1, offset, isAdmin)
	if err != nil {
		s.log.Error("can't get posts from db", slog.Any("error", err))

		http.Error(w, "can't get posts from db", http.StatusInternalServerError)

		return
	}

	tmplData := PostsPageData{
		Posts:        posts,
		IsAdmin:      isAdmin,
		HasNextPages: false,
		NextOffset:   offset + len(posts),
	}

	if len(posts) == s.pageLimit+1 {
		tmplData.HasNextPages = true
		tmplData.Posts = tmplData.Posts[:len(tmplData.Posts)-1]
		tmplData.NextOffset = offset + len(tmplData.Posts)
	}

	s.renderTemplate(w, "pagination-posts", tmplData)
}

func (s *Server) postPage(w http.ResponseWriter, r *http.Request) {
	isAdmin := r.Context().Value(middleware.IsAdminKey) != nil

	slug := r.PathValue("slug")

	post, err := s.Blog.PostBySlug(r.Context(), slug)
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
		Slug        string
		CreatedAt   string
		UpdatedAt   string
		IsPublished bool
		IsAdmin     bool
	}{
		ID:          post.ID,
		Title:       post.Title,
		Description: post.Description,
		Content:     tmplContent,
		Slug:        post.Slug,
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
	slug := r.FormValue("slug")
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
		Slug:        slug,
		Description: description,
		Filename:    header.Filename,
		Content:     content,
		IsPublished: false,
	}

	post, err := s.Blog.CreatePost(r.Context(), postParms)
	if err != nil {
		s.log.Error("can't create post", slog.Any("error", err))

		http.Error(w, "can't create post", http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, "/blog/posts/"+post.Slug, http.StatusFound)
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
	slug := r.URL.Query().Get("slug")

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

	w.Header().Set("HX-Redirect", "/blog/posts/"+slug)
	w.WriteHeader(http.StatusOK)
}
