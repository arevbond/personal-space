package server

import (
	"log/slog"
	"net/http"

	"github.com/arevbond/arevbond-blog/internal/blog/domain"
)

func (s *Server) registerBlogRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /blog/posts", s.htmlPosts)
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
