package server

import (
	"log/slog"
	"net/http"
)

func (s *Server) renderTemplate(w http.ResponseWriter, templateName string, data any) {
	s.log.Debug("render template", slog.String("name", templateName))

	if err := s.tmpl.ExecuteTemplate(w, templateName, data); err != nil {
		s.log.Error("can't render template",
			slog.String("template name", templateName),
			slog.Any("error", err))

		http.Error(w, "can't execute template", http.StatusInternalServerError)

		return
	}
}

func (s *Server) renderError(w http.ResponseWriter, errorMsg string, err error, statusCode int) {
	s.log.Error(errorMsg, slog.Any("error", err))

	http.Error(w, errorMsg, statusCode)
}
