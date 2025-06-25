package server

import (
	"log/slog"
	"net/http"
)

type Auth interface {
	IsAdminToken(token string) bool
}

func (s *Server) registerAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /login-admin", s.loginPage)
	mux.HandleFunc("POST /verify-token", s.verifyAdminToken)
}

func (s *Server) loginPage(w http.ResponseWriter, r *http.Request) {
	s.renderTemplate(w, "admin_login.html", nil)
}

func (s *Server) verifyAdminToken(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "can't parse form", http.StatusBadRequest)

		return
	}

	incomeToken := r.Form.Get("token")

	s.log.Debug("income admin token", slog.String("token", incomeToken))

	if s.Auth.IsAdminToken(incomeToken) {
		setTokenCookie(w, "JWT_TOKEN")
		s.log.Debug("success set cookie")
	} else {
		s.log.Debug("invalid income token")
	}
}

func setTokenCookie(w http.ResponseWriter, token string) {
	// время жизни куки в секундах
	const ttl = 3600

	//nolint: exhaustruct // default cookie struct
	cookie := http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		MaxAge:   ttl,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)
}
