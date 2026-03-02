package connections

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMySQLConnectionConfig_WithURL(t *testing.T) {
	cfg, err := NewMySQLConnectionConfig(map[string]any{
		"url": "mysql://db-user:db-pass@localhost:3306/app_db?tls=true&parseTime=true",
	})
	require.NoError(t, err)

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 3306, cfg.Port)
	assert.Equal(t, "db-user", cfg.User)
	assert.Equal(t, "db-pass", cfg.Password)
	assert.Equal(t, "app_db", cfg.Database)
	assert.Equal(t, "true", cfg.Params["tls"])
	assert.Equal(t, "true", cfg.Params["parseTime"])
}

func TestNewMySQLConnectionConfig_WithLegacyFields(t *testing.T) {
	cfg, err := NewMySQLConnectionConfig(map[string]any{
		"host":     "localhost",
		"port":     3306,
		"user":     "root",
		"password": "secret",
		"database": "app_db",
	})
	require.NoError(t, err)

	assert.Empty(t, cfg.URL)
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 3306, cfg.Port)
	assert.Equal(t, "root", cfg.User)
	assert.Equal(t, "secret", cfg.Password)
	assert.Equal(t, "app_db", cfg.Database)
}

func TestNewMySQLConnectionConfig_InvalidURL(t *testing.T) {
	_, err := NewMySQLConnectionConfig(map[string]any{
		"url": "postgres://db-user:db-pass@localhost:5432/app_db",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid mysql connection url scheme")
}
