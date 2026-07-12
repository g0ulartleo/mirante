package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitRepoCommandScaffoldsNodeRuntime(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "runtime")
	cmd := &InitRepoCommand{}

	err := cmd.Run([]string{"--runtime", "nodejs", "--dir", dir})
	require.NoError(t, err)

	assertFileExists(t, filepath.Join(dir, "package.json"))
	assertFileExists(t, filepath.Join(dir, "tsconfig.json"))
	assertFileExists(t, filepath.Join(dir, "src/server.ts"))
	assertFileExists(t, filepath.Join(dir, "src/alarms/check-server-count.ts"))
	assertFileExists(t, filepath.Join(dir, ".env.example"))
	assertFileExists(t, filepath.Join(dir, "README.md"))
	assertFileExists(t, filepath.Join(dir, ".gitignore"))
	assertFileExists(t, filepath.Join(dir, ".dockerignore"))
	assertFileExists(t, filepath.Join(dir, "docker-compose.yml"))
	assertFileExists(t, filepath.Join(dir, "Dockerfile"))
	assertFileExists(t, filepath.Join(dir, "mirante.yaml"))
	marker := readFile(t, filepath.Join(dir, "mirante.runtime.yaml"))
	assert.Contains(t, marker, "runtime: nodejs")
	assert.Contains(t, marker, "alarms_dir: src/alarms")
	miranteCfg := readFile(t, filepath.Join(dir, "mirante.yaml"))
	assert.Contains(t, miranteCfg, "runtime:50051")
	dockerfile := readFile(t, filepath.Join(dir, "Dockerfile"))
	assert.Contains(t, dockerfile, "FROM node:22-alpine AS deps")
	assert.Contains(t, dockerfile, "FROM node:22-alpine AS prod")
	assert.Contains(t, dockerfile, "npm ci --omit=dev")
	assert.Contains(t, dockerfile, `CMD ["node", "dist/server.js"]`)
	dockerIgnore := readFile(t, filepath.Join(dir, ".dockerignore"))
	assert.Contains(t, dockerIgnore, ".env.*")
	assert.Contains(t, dockerIgnore, "terraform/")
	compose := readFile(t, filepath.Join(dir, "docker-compose.yml"))
	assert.Contains(t, compose, "target: dev")
}

func TestInitRepoCommandScaffoldsGoRuntime(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "runtime")
	cmd := &InitRepoCommand{}

	err := cmd.Run([]string{"--runtime", "go", "--dir", dir})
	require.NoError(t, err)

	assertFileExists(t, filepath.Join(dir, "go.mod"))
	assertFileExists(t, filepath.Join(dir, "cmd/runtime/main.go"))
	assertFileExists(t, filepath.Join(dir, "internal/alarms/check_server_count.go"))
	assertFileExists(t, filepath.Join(dir, ".env.example"))
	assertFileExists(t, filepath.Join(dir, "README.md"))
	assertFileExists(t, filepath.Join(dir, ".gitignore"))
	assertFileExists(t, filepath.Join(dir, ".dockerignore"))
	assertFileExists(t, filepath.Join(dir, "docker-compose.yml"))
	assertFileExists(t, filepath.Join(dir, "Dockerfile"))
	assertFileExists(t, filepath.Join(dir, "mirante.yaml"))
	marker := readFile(t, filepath.Join(dir, "mirante.runtime.yaml"))
	assert.Contains(t, marker, "runtime: go")
	assert.Contains(t, marker, "alarms_dir: internal/alarms")
	main := readFile(t, filepath.Join(dir, "cmd/runtime/main.go"))
	assert.Contains(t, main, "alarmsdk.ServeRuntime")
	alarm := readFile(t, filepath.Join(dir, "internal/alarms/check_server_count.go"))
	assert.Contains(t, alarm, "PING_URL")
	assert.Contains(t, alarm, "alarmsdk.Healthy")
	miranteCfg := readFile(t, filepath.Join(dir, "mirante.yaml"))
	assert.Contains(t, miranteCfg, "runtime:50051")
}

func TestNewAlarmCommandCreatesNodeAlarmFromMarker(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "mirante.runtime.yaml"), []byte("runtime: nodejs\nalarms_dir: src/alarms\n"), 0644))
	withWorkingDir(t, dir)

	cmd := &NewAlarmCommand{}
	err := cmd.Run([]string{"server-events-dlq"})
	require.NoError(t, err)

	content := readFile(t, filepath.Join(dir, "src/alarms/server-events-dlq.ts"))
	assert.Contains(t, content, "export const serverEventsDlq")
	assert.Contains(t, content, `id: "server-events-dlq"`)
}

func TestNewAlarmCommandCreatesGoAlarmFromMarker(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "mirante.runtime.yaml"), []byte("runtime: go\nalarms_dir: internal/alarms\n"), 0644))
	withWorkingDir(t, dir)

	cmd := &NewAlarmCommand{}
	err := cmd.Run([]string{"server-events-dlq"})
	require.NoError(t, err)

	content := readFile(t, filepath.Join(dir, "internal/alarms/server_events_dlq.go"))
	assert.Contains(t, content, `ID:          "server-events-dlq"`)
	assert.Contains(t, content, "func runServerEventsDlq")
}

func TestNewAlarmCommandFailsOutsideRuntimeRepo(t *testing.T) {
	withWorkingDir(t, t.TempDir())
	cmd := &NewAlarmCommand{}

	err := cmd.Run([]string{"server-events-dlq"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mirante.runtime.yaml not found")
}

func TestNewAlarmCommandRejectsInvalidID(t *testing.T) {
	cmd := &NewAlarmCommand{}
	err := cmd.Run([]string{"ServerEvents"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	require.NoError(t, err)
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
}

func withWorkingDir(t *testing.T, dir string) {
	t.Helper()
	previous, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(previous))
	})
}
