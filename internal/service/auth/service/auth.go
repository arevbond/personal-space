package service

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrUnexpectedSigningMethod = errors.New("unexpected signing method")

type Auth struct {
	log          *slog.Logger
	adminToken   string
	secretKeyJWT string
}

func New(log *slog.Logger, adminToken string, secretKey string) *Auth {
	return &Auth{log: log, adminToken: adminToken, secretKeyJWT: secretKey}
}

func (a *Auth) IsAdminToken(token string) bool {
	return a.adminToken == token
}

func (a *Auth) NewJWT() (string, error) {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(a.secretKeyJWT))
	if err != nil {
		return "", fmt.Errorf("can't sign string: %w", err)
	}

	return signedToken, nil
}

func (a *Auth) VerifyJWT(tokenStr string) (bool, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedSigningMethod, token.Header["alg"])
		}

		return []byte(a.secretKeyJWT), nil
	})
	if err != nil {
		return false, fmt.Errorf("can't parse jwt: %w", err)
	}

	return token.Valid, nil
}
