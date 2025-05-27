package server

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/arevbond/arevbond-blog/internal/blog/domain"
)

func (s *Server) registerBlogRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /blog/posts", s.htmlPosts)
	mux.HandleFunc("GET /blog/post/{id}", s.htmlPost)
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
