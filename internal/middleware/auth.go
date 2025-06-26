package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
)

type Auth interface {
	VerifyJWT(tokenStr string) (bool, error)
}

type contextKey string

const IsAdminKey contextKey = "is-admin"

func RequireAuth(auth Auth, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		authHandler := func(w http.ResponseWriter, r *http.Request) {
			token, err := r.Cookie("token")
			if err != nil {
				switch {
				case errors.Is(err, http.ErrNoCookie):
					http.Error(w, "cookie not found", http.StatusBadRequest)
				default:
					log.Error("can't get cookie", slog.Any("error", err))
					http.Error(w, "server error", http.StatusInternalServerError)
				}

				return
			}

			isValid, err := auth.VerifyJWT(token.Value)
			if err != nil {
				log.Error("can't verify jwt", slog.Any("error", err))
				http.Error(w, "server error", http.StatusInternalServerError)

				return
			}

			if !isValid {
				http.Error(w, "access denied", http.StatusForbidden)

				return
			}

			// Only admin can create JWTs, so any valid JWT = admin user
			newCtx := context.WithValue(r.Context(), IsAdminKey, true)

			next.ServeHTTP(w, r.WithContext(newCtx))
		}

		return http.HandlerFunc(authHandler)
	}
}

func OptionalAuth(auth Auth, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		authHandler := func(w http.ResponseWriter, r *http.Request) {
			token, err := r.Cookie("token")
			if err != nil {
				next.ServeHTTP(w, r)

				return
			}

			isValid, err := auth.VerifyJWT(token.Value)
			if err != nil {
				log.Error("can't verify jwt", slog.Any("error", err))
				next.ServeHTTP(w, r)

				return
			}

			if isValid {
				newCtx := context.WithValue(r.Context(), IsAdminKey, true)
				next.ServeHTTP(w, r.WithContext(newCtx))

				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(authHandler)
	}
}
