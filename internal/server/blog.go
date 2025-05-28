package server

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/arevbond/arevbond-blog/internal/blog/domain"
)

type Blog interface {
	Posts(ctx context.Context, limit, offset int) ([]*domain.Post, error)
	Post(ctx context.Context, id int) (*domain.Post, error)
	CreatePost(ctx context.Context, params domain.PostParams) (int, error)
}

func (s *Server) registerBlogRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /blog/posts", s.htmlPosts)
	mux.HandleFunc("GET /blog/post/{id}", s.htmlPost)

	mux.HandleFunc("GET /blog/post/form", s.htmlCreatePost)
	mux.HandleFunc("POST /blog/post", s.createPost)
}

func (s *Server) htmlPosts(w http.ResponseWriter, r *http.Request) {
	const pageLimit = 10

	posts, err := s.Blog.Posts(r.Context(), pageLimit, 0)
	if err != nil {
		http.Error(w, "can't get posts from db", http.StatusInternalServerError)

		return
	}

	tmplData := struct {
		Posts []*domain.Post
	}{
		Posts: posts,
	}

	if err = s.tmpl.ExecuteTemplate(w, "posts.html", tmplData); err != nil {
		s.log.Error("htmlPosts", slog.Any("error", err))
		http.Error(w, "can't render posts.html", http.StatusInternalServerError)

		return
	}
}

func (s *Server) htmlPost(w http.ResponseWriter, r *http.Request) {
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

	w.WriteHeader(http.StatusOK)

	s.renderTemplate(w, "post.html", post)
}

func (s *Server) htmlCreatePost(w http.ResponseWriter, r *http.Request) {
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
	}

	postID, err := s.Blog.CreatePost(r.Context(), postParms)
	if err != nil {
		s.log.Error("can't create post", slog.Any("error", err))

		http.Error(w, "can't create post", http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/blog/post/%d", postID), http.StatusFound)
}
