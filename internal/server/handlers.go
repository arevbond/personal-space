package server

import (
	"log/slog"
	"net/http"

	"github.com/arevbond/arevbond-blog/internal/models"
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

func (s *Server) htmlCVpreview(w http.ResponseWriter, r *http.Request) {
	if err := s.tmpl.ExecuteTemplate(w, "edit_cv.html", nil); err != nil {
		s.log.Error("can't render edit cv html", slog.Any("error", err))
		http.Error(w, "Error while render page", http.StatusInternalServerError)

		return
	}
}

func (s *Server) htmlAllCV(w http.ResponseWriter, r *http.Request) {
	cvs, err := s.manager.ListCV(r.Context())
	if err != nil {
		s.log.Error("can't get all cv from db", slog.Any("error", err))
		http.Error(w, "Error while process db request", http.StatusInternalServerError)

		return
	}

	tmplData := struct {
		ListCV []models.CV
	}{
		ListCV: cvs,
	}

	if err = s.tmpl.ExecuteTemplate(w, "all_cv.html", tmplData); err != nil {
		s.log.Error("can't render all cv", slog.Any("error", err))
		http.Error(w, "Error while render page", http.StatusInternalServerError)

		return
	}
}
