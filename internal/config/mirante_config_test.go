package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadAuthConfigDoesNotReadAuthYAML(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "config"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config", "auth.yaml"), []byte(`api_key: from-auth-yaml`), 0644))
	mirantePath := filepath.Join(dir, "config", "mirante.yaml")
	require.NoError(t, os.WriteFile(mirantePath, []byte(`auth:
  api_key: from-mirante
`), 0644))
	t.Setenv("MIRANTE_CONFIG", mirantePath)

	cfg, err := LoadMiranteConfig()
	require.NoError(t, err)
	assert.Equal(t, "from-mirante", cfg.Auth.APIKey)
}

func TestLoadMiranteConfigExpandsEnvAndLoadsRuntimes(t *testing.T) {
	t.Setenv("API_KEY", "secret")
	dir := t.TempDir()
	path := filepath.Join(dir, "mirante.yaml")
	require.NoError(t, os.WriteFile(path, []byte(`
storage:
  driver: redis
alarm_runtime:
  timeout: 5s
  runtimes:
    go:
      addr: go-alarm-runtime:50051
    nodejs:
      addr: nodejs-alarm-runtime:50051
auth:
  api_key: ${API_KEY}
`), 0644))
	t.Setenv("MIRANTE_CONFIG", path)

	cfg, err := LoadMiranteConfig()
	require.NoError(t, err)

	assert.Equal(t, "redis", cfg.Storage.Driver)
	assert.Equal(t, "5s", cfg.AlarmRuntime.Timeout)
	assert.Equal(t, "go-alarm-runtime:50051", cfg.AlarmRuntime.Runtimes["go"].Addr)
	assert.Equal(t, "nodejs-alarm-runtime:50051", cfg.AlarmRuntime.Runtimes["nodejs"].Addr)
	assert.Equal(t, "secret", cfg.Auth.APIKey)
}

func TestLoadMiranteConfigAppliesAlarmRuntimeEnvOverrides(t *testing.T) {
	t.Setenv("ALARM_RUNTIME_ADDR", "env-alarm-runtime:50051")
	t.Setenv("ALARM_RUNTIME_TIMEOUT", "9s")
	t.Setenv("MIRANTE_CONFIG", filepath.Join(t.TempDir(), "missing.yaml"))

	cfg, err := LoadMiranteConfig()
	require.NoError(t, err)

	assert.Equal(t, "9s", cfg.AlarmRuntime.Timeout)
	assert.Equal(t, "env-alarm-runtime:50051", cfg.AlarmRuntime.Runtimes["env"].Addr)
}
