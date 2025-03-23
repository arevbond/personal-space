package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_New(t *testing.T) {
	t.Run("success read config", func(t *testing.T) {
		dir := t.TempDir()
		file, err := os.CreateTemp(dir, "application.yaml")
		require.NoError(t, err)

		_, err = file.Write([]byte(`
server:
  host: "app-arevbond"
  port: 8080
`))
		require.NoError(t, err)

		cfg, err := New(file.Name())
		assert.NoError(t, err)

		assert.Equal(t, "app-arevbond", cfg.Server.Host)
		assert.Equal(t, 8080, cfg.Server.Port)
	})
	t.Run("failed unmarshall yaml config", func(t *testing.T) {
		dir := t.TempDir()
		file, err := os.CreateTemp(dir, "application.yaml")
		require.NoError(t, err)

		_, err = file.Write([]byte(`
{
"server":
  "host": "app-arevbond"
  "port": 8080
}
`))
		require.NoError(t, err)

		cfg, err := New(file.Name())
		assert.Error(t, err)
		assert.Empty(t, cfg)
	})
	t.Run("config file doesn't exists", func(t *testing.T) {
		dir := t.TempDir()

		cfg, err := New(dir + "file123321")
		assert.Error(t, err)
		assert.Empty(t, cfg)
	})
}
