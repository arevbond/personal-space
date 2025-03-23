package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/arevbond/arevbond-blog/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_Run(t *testing.T) {
	srv := New(slog.Default(), config.Server{
		Host: "localhost",
		Port: 9988,
	}).WithRoutes()
	assert.NotNil(t, srv.Server)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srvFinished := make(chan struct{})

	go func() {
		err := srv.Run(ctx)
		assert.NoError(t, err)

		close(srvFinished)
	}()

	// need for launch server
	time.Sleep(1e2 * time.Millisecond)

	resp, err := http.Get("http://localhost:9988/ping")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, "pong", string(data))
	cancel()

	<-srvFinished
}
