package server

import (
	"log/slog"
	"net/http"
)

func (s *Server) ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("pong"))
}

func (s *Server) htmlIndex(w http.ResponseWriter, r *http.Request) {
	if err := s.tmpl.ExecuteTemplate(w, "index.html", nil); err != nil {
		s.log.Error("can't render index html", slog.Any("error", err))
		http.Error(w, "Error while render page", http.StatusInternalServerError)

		return
	}
}

func (s *Server) htmlPosts(w http.ResponseWriter, r *http.Request) {
	if err := s.tmpl.ExecuteTemplate(w, "posts.html", nil); err != nil {
		http.Error(w, "Error while rendering posts", http.StatusInternalServerError)

		return
	}
}
