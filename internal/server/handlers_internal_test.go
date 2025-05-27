package server

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arevbond/arevbond-blog/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestHandlers_ping(t *testing.T) {
	t.Parallel()

	srv := New(slog.Default(), config.Server{
		Host: "",
		Port: 0,
	}, Services{})
	handler := http.HandlerFunc(srv.ping)
	req := httptest.NewRequest(http.MethodGet, "/ping", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	body := rr.Body.String()
	assert.Equal(t, "pong", body)
}
