package connections

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPostgresConnectionConfig_WithURL(t *testing.T) {
	cfg, err := NewPostgresConnectionConfig(map[string]any{
		"url":         "postgres://db-user:db-pass@localhost:5432/app_db?sslmode=require",
		"sslverify":   true,
		"sslrootcert": "/tmp/rds-ca.pem",
	})
	require.NoError(t, err)

	assert.Equal(t, "postgres://db-user:db-pass@localhost:5432/app_db?sslmode=require", cfg.URL)
	assert.Equal(t, "/tmp/rds-ca.pem", cfg.SSLRootCert)
	assert.True(t, cfg.SSLVerify)
}

func TestNewPostgresConnectionConfig_WithLegacyFields(t *testing.T) {
	cfg, err := NewPostgresConnectionConfig(map[string]any{
		"host":     "localhost",
		"port":     5432,
		"user":     "postgres",
		"password": "secret",
		"database": "app_db",
	})
	require.NoError(t, err)

	assert.Empty(t, cfg.URL)
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 5432, cfg.Port)
	assert.Equal(t, "postgres", cfg.User)
	assert.Equal(t, "secret", cfg.Password)
	assert.Equal(t, "app_db", cfg.Database)
}

func TestNewPostgresConnectionConfig_InvalidURL(t *testing.T) {
	_, err := NewPostgresConnectionConfig(map[string]any{
		"url": "mysql://db-user:db-pass@localhost:5432/app_db",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid postgres connection url scheme")
}
