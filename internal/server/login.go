package server

import "net/http"

func (s *Server) registerAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /login-admin", s.htmlLogin)
}

func (s *Server) htmlLogin(w http.ResponseWriter, r *http.Request) {
	s.renderTemplate(w, "admin_login.html", nil)
}
