package server_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/arevbond/arevbond-blog/internal/config"
	"github.com/arevbond/arevbond-blog/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_Run(t *testing.T) {
	t.Parallel()

	srv := server.New(slog.Default(), config.Server{
		Host: "localhost",
		Port: 9988,
	}, server.Config{}).WithRoutes()
	assert.NotNil(t, srv.Server)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	srvFinished := make(chan struct{})

	go func() {
		err := srv.Run(ctx)
		assert.NoError(t, err)

		close(srvFinished)
	}()

	// need for launch server
	time.Sleep(1e2 * time.Millisecond)

	//nolint: noctx // request for test
	resp, err := http.Get("http://localhost:9988/ping")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, "pong", string(data))
	cancel()

	<-srvFinished
}
