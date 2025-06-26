package server

import (
	"log/slog"
	"net/http"
)

type Auth interface {
	IsAdminToken(token string) bool
	NewJWT() (string, error)
	VerifyJWT(tokenStr string) (bool, error)
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

	if s.Auth.IsAdminToken(incomeToken) {
		var token string
		token, err = s.Auth.NewJWT()

		if err != nil {
			s.log.Error("can't create jwt", slog.Any("error", err))

			http.Error(w, "can't create jwt", http.StatusInternalServerError)

			return
		}

		setTokenCookie(w, token)

		s.log.Debug("success set cookie")

		s.renderTemplate(w, "success_admin_login", nil)
	}
}

func setTokenCookie(w http.ResponseWriter, token string) {
	const TokenTTL = 3600 // 1 hour

	//nolint: exhaustruct // default cookie struct
	cookie := http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		MaxAge:   TokenTTL,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)
}
