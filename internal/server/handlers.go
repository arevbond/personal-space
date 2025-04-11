package server

import (
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

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

func (s *Server) htmlCVedit(w http.ResponseWriter, r *http.Request) {
	s.log.Debug("cv edit", slog.String("id", r.PathValue("id")))

	if err := s.tmpl.ExecuteTemplate(w, "edit_cv.html", nil); err != nil {
		s.log.Error("can't render edit cv html", slog.Any("error", err))
		http.Error(w, "Error while render page", http.StatusInternalServerError)

		return
	}
}

func (s *Server) htmlResumeList(w http.ResponseWriter, r *http.Request) {
	cvs, err := s.manager.Resumes(r.Context())
	if err != nil {
		s.log.Error("can't get all cv from db", slog.Any("error", err))
		http.Error(w, "Error while process db request", http.StatusInternalServerError)

		return
	}

	tmplData := struct {
		ListCV []models.Resume
	}{
		ListCV: cvs,
	}

	if err = s.tmpl.ExecuteTemplate(w, "all_cv.html", tmplData); err != nil {
		s.log.Error("can't render all cv", slog.Any("error", err))
		http.Error(w, "Error while render page", http.StatusInternalServerError)

		return
	}
}

func (s *Server) uploadResume(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error while parsing form", http.StatusInternalServerError)

		return
	}

	file, heads, err := r.FormFile("cv")
	if err != nil {
		http.Error(w, "Error while parsing file", http.StatusInternalServerError)

		return
	}

	s.log.Debug("income cv", slog.String("name", heads.Filename))

	strs := strings.Split(heads.Filename, ".")

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error while reading file", http.StatusInternalServerError)

		return
	}

	resume := models.Resume{
		ID:            -1,
		Name:          heads.Filename,
		Content:       data,
		FileExtension: strs[len(strs)-1],
		UpdatedAt:     time.Now(),
	}

	err = s.manager.UploadResume(r.Context(), resume)
	if err != nil {
		http.Error(w, "Error while upload cv", http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, "/cv", http.StatusMovedPermanently)
}
