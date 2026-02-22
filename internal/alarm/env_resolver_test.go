package alarm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveStringEnvVars(t *testing.T) {
	t.Setenv("HOST", "example.com")
	t.Setenv("PORT", "443")

	resolved, err := resolveStringEnvVars("https://${HOST}:${PORT}/health")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com:443/health", resolved)
}

func TestResolveStringEnvVars_UnterminatedPlaceholder(t *testing.T) {
	_, err := resolveStringEnvVars("https://${HOST/health")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unterminated environment variable placeholder")
}

func TestResolveStringEnvVars_InvalidVariableName(t *testing.T) {
	_, err := resolveStringEnvVars("value=${1HOST}")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid environment variable name: 1HOST")
}
