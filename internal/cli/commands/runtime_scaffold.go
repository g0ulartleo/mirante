package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/g0ulartleo/mirante/internal/cli"
	"gopkg.in/yaml.v3"
)

type RuntimeMarker struct {
	Runtime   string `yaml:"runtime"`
	AlarmsDir string `yaml:"alarms_dir"`
}

type InitRepoCommand struct{}

func (c *InitRepoCommand) Name() string {
	return "init repo"
}

func (c *InitRepoCommand) Description() string {
	return "Scaffold an alarm runtime repository"
}

func (c *InitRepoCommand) Usage() string {
	return "init repo --runtime <nodejs|go> --dir <path>"
}

func (c *InitRepoCommand) Run(args []string) error {
	runtime, dir, err := parseInitRepoArgs(args)
	if err != nil {
		return err
	}

	switch runtime {
	case "nodejs":
		return scaffoldNodeRepo(dir)
	case "go":
		return scaffoldGoRepo(dir)
	default:
		return fmt.Errorf("unsupported runtime %q; expected nodejs or go", runtime)
	}
}

type NewAlarmCommand struct{}

func (c *NewAlarmCommand) Name() string {
	return "new alarm"
}

func (c *NewAlarmCommand) Description() string {
	return "Create an alarm file inside a runtime repository"
}

func (c *NewAlarmCommand) Usage() string {
	return "new alarm <alarm-id>"
}

func (c *NewAlarmCommand) Run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: mirante %s", c.Usage())
	}

	alarmID := args[0]
	if err := validateAlarmID(alarmID); err != nil {
		return err
	}

	marker, err := loadRuntimeMarker("mirante.runtime.yaml")
	if err != nil {
		return err
	}

	switch marker.Runtime {
	case "nodejs":
		path := filepath.Join(marker.AlarmsDir, alarmID+".ts")
		return writeFileExclusive(path, nodeAlarmTemplate(alarmID))
	case "go":
		path := filepath.Join(marker.AlarmsDir, strings.ReplaceAll(alarmID, "-", "_")+".go")
		return writeFileExclusive(path, goAlarmTemplate(alarmID))
	default:
		return fmt.Errorf("unsupported runtime %q in mirante.runtime.yaml", marker.Runtime)
	}
}

func parseInitRepoArgs(args []string) (string, string, error) {
	var runtime string
	var dir string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--runtime":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("--runtime requires a value")
			}
			runtime = args[i+1]
			i++
		case "--dir":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("--dir requires a value")
			}
			dir = args[i+1]
			i++
		default:
			return "", "", fmt.Errorf("unknown argument %q; usage: mirante init repo --runtime <nodejs|go> --dir <path>", args[i])
		}
	}
	if runtime == "" || dir == "" {
		return "", "", fmt.Errorf("usage: mirante init repo --runtime <nodejs|go> --dir <path>")
	}
	return runtime, dir, nil
}

func scaffoldNodeRepo(dir string) error {
	marker := RuntimeMarker{Runtime: "nodejs", AlarmsDir: "src/alarms"}
	files := map[string]string{
		"package.json":                     nodePackageJSON(),
		"tsconfig.json":                    nodeTSConfig(),
		"src/server.ts":                    nodeServerTemplate(),
		"src/alarms/check-server-count.ts": nodeAlarmTemplate("check-server-count"),
		".env.example":                     nodeEnvExample(),
		"mirante.runtime.yaml":             markerYAML(marker),
		"README.md":                        runtimeReadme(),
		".gitignore":                       nodeGitIgnore(),
		"docker-compose.yml":               nodeDockerCompose(),
		"Dockerfile":                       nodeDockerfile(),
		"mirante.yaml":                     miranteConfig(),
	}
	return writeScaffoldFiles(dir, files)
}

func scaffoldGoRepo(dir string) error {
	marker := RuntimeMarker{Runtime: "go", AlarmsDir: "internal/alarms"}
	files := map[string]string{
		"go.mod":                                goRuntimeMod(),
		"cmd/runtime/main.go":                   goRuntimeMain(),
		"internal/alarms/check_server_count.go": goAlarmTemplate("check-server-count"),
		".env.example":                          goEnvExample(),
		"mirante.runtime.yaml":                  markerYAML(marker),
		"README.md":                             runtimeReadme(),
		".gitignore":                            goGitIgnore(),
	}
	return writeScaffoldFiles(dir, files)
}

func writeScaffoldFiles(dir string, files map[string]string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create runtime directory: %w", err)
	}
	for path, content := range files {
		if err := writeFileExclusive(filepath.Join(dir, path), content); err != nil {
			return err
		}
	}
	return nil
}

func writeFileExclusive(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", path, err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("file already exists: %s", path)
		}
		return fmt.Errorf("failed to create %s: %w", path, err)
	}
	defer file.Close()
	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}

func loadRuntimeMarker(path string) (*RuntimeMarker, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("mirante.runtime.yaml not found; run `mirante init repo --runtime <nodejs|go> --dir <path>` first")
		}
		return nil, fmt.Errorf("failed to read mirante.runtime.yaml: %w", err)
	}
	var marker RuntimeMarker
	if err := yaml.Unmarshal(data, &marker); err != nil {
		return nil, fmt.Errorf("failed to parse mirante.runtime.yaml: %w", err)
	}
	if marker.Runtime == "" || marker.AlarmsDir == "" {
		return nil, fmt.Errorf("mirante.runtime.yaml must include runtime and alarms_dir")
	}
	return &marker, nil
}

func markerYAML(marker RuntimeMarker) string {
	return fmt.Sprintf("runtime: %s\nalarms_dir: %s\n", marker.Runtime, marker.AlarmsDir)
}

func validateAlarmID(id string) error {
	if id == "" {
		return fmt.Errorf("alarm id is required")
	}
	matched, err := regexp.MatchString(`^[a-z0-9][a-z0-9-]*$`, id)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("alarm id %q is invalid; use lowercase letters, numbers, and dashes", id)
	}
	return nil
}

func nodePackageJSON() string {
	return `{
  "name": "mirante-alarm-runtime",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "scripts": {
    "start": "tsx src/server.ts"
  },
  "dependencies": {
    "@mirante/alarms-sdk": "0.1.0"
  },
  "devDependencies": {
    "tsx": "^4.19.0",
    "typescript": "^5.6.0"
  }
}
`
}

func nodeTSConfig() string {
	return `{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "strict": true,
    "skipLibCheck": true
  },
  "include": ["src/**/*.ts"]
}
`
}

func nodeServerTemplate() string {
	return `import { serveRuntime } from "@mirante/alarms-sdk";

await serveRuntime({
  alarmsDir: new URL("./alarms", import.meta.url).pathname,
  addr: process.env.ALARM_RUNTIME_ADDR ?? "127.0.0.1:50051",
});
`
}

func nodeAlarmTemplate(alarmID string) string {
	className := toPascalCase(alarmID)
	exportName := strings.ToLower(className[:1]) + className[1:]
	return fmt.Sprintf(`import { healthy } from "@mirante/alarms-sdk";

export const %s = {
  id: %q,
  name: %q,
  description: "Describe what this alarm checks.",
  howToFix: "Describe how to fix failures.",
  interval: "1m",
  notifications: {
    slackWebhooks: async () => [],
    emails: async () => [],
  },
  async run() {
    return healthy("OK");
  },
};
`, exportName, alarmID, humanizeAlarmID(alarmID))
}

func nodeEnvExample() string {
	return `ALARM_RUNTIME_ADDR=0.0.0.0:50051

# Mirante Core
API_KEY=
DASHBOARD_BASIC_AUTH_USERNAME=admin
DASHBOARD_BASIC_AUTH_PASSWORD=
`
}

func goRuntimeMod() string {
	return `module mirante-alarm-runtime

go 1.23.6
`
}

func goRuntimeMain() string {
	return `package main

func main() {
    // TODO: wire generated alarm runtime server.
}
`
}

func goAlarmTemplate(alarmID string) string {
	return fmt.Sprintf(`package alarms

// AlarmID identifies this runtime-owned alarm.
const AlarmID = %q
`, alarmID)
}

func goEnvExample() string {
	return "ALARM_RUNTIME_ADDR=127.0.0.1:50051\n"
}

func nodeDockerCompose() string {
	return `services:
  redis:
    image: redis:7.2-alpine
    ports:
      - 6379:6379
    volumes:
      - redis-data:/data

  mirante:
    image: g0ulartleo/mirante:latest
    ports:
      - 40169:40169
    env_file: .env
    environment:
      REDIS_ADDR: redis:6379
      HTTP_ADDR: 0.0.0.0
      MIRANTE_CONFIG: /etc/mirante/mirante.yaml
    volumes:
      - ./mirante.yaml:/etc/mirante/mirante.yaml
    depends_on:
      redis:
        condition: service_started
      runtime:
        condition: service_healthy
    extra_hosts:
      - host.docker.internal:host-gateway

  worker:
    image: g0ulartleo/mirante:latest
    command: ["./bin/worker"]
    env_file: .env
    environment:
      REDIS_ADDR: redis:6379
      MIRANTE_CONFIG: /etc/mirante/mirante.yaml
    volumes:
      - ./mirante.yaml:/etc/mirante/mirante.yaml
      - ~/.aws:/home/node/.aws:ro
    depends_on:
      redis:
        condition: service_started
      runtime:
        condition: service_healthy
    extra_hosts:
      - host.docker.internal:host-gateway

  scheduler:
    image: g0ulartleo/mirante:latest
    command: ["./bin/scheduler"]
    env_file: .env
    environment:
      REDIS_ADDR: redis:6379
      MIRANTE_CONFIG: /etc/mirante/mirante.yaml
    volumes:
      - ./mirante.yaml:/etc/mirante/mirante.yaml
    depends_on:
      redis:
        condition: service_started
      runtime:
        condition: service_healthy

  runtime:
    build: .
    ports:
      - 50051:50051
    env_file: .env
    environment:
      ALARM_RUNTIME_ADDR: 0.0.0.0:50051
      AWS_SHARED_CREDENTIALS_FILE: /home/node/.aws/credentials
      AWS_CONFIG_FILE: /home/node/.aws/config
    volumes:
      - ./src:/app/src
      - ~/.aws:/home/node/.aws:ro
    healthcheck:
      test: ["CMD", "node", "-e", "require('net').createConnection(50051).on('connect',()=>process.exit(0)).on('error',()=>process.exit(1))"]
      interval: 2s
      timeout: 3s
      retries: 10
      start_period: 10s
    depends_on:
      - mirante

volumes:
  redis-data:
`
}

func nodeDockerfile() string {
	return `FROM node:22-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
EXPOSE 50051
CMD ["npm", "start"]
`
}

func miranteConfig() string {
	return `storage:
  driver: redis

redis:
  addr: redis:6379

http:
  addr: 0.0.0.0
  port: "40169"

alarm_runtime:
  timeout: 30s
  runtimes:
    runtime:
      addr: runtime:50051

auth:
  api_key: ${API_KEY}
  basic:
    username: ${DASHBOARD_BASIC_AUTH_USERNAME}
    password: ${DASHBOARD_BASIC_AUTH_PASSWORD}
`
}

func nodeGitIgnore() string {
	return `node_modules/
.env
.DS_Store
`
}

func goGitIgnore() string {
	return `.env
.DS_Store
tmp/
dist/
`
}

func runtimeReadme() string {
	return fmt.Sprintf(`# Mirante Alarms Runtime

Generated by `+"`"+`mirante init repo`+"`"+`.

This repository owns alarm definitions and implementations.
Mirante core syncs alarm metadata from this server application and use it to execute alarms.
`)
}

func toPascalCase(id string) string {
	parts := strings.Split(id, "-")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}

func humanizeAlarmID(id string) string {
	parts := strings.Split(id, "-")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func init() {
	initRepo := &InitRepoCommand{}
	cli.RegisterCommand(initRepo.Name(), initRepo)

	newAlarm := &NewAlarmCommand{}
	cli.RegisterCommand(newAlarm.Name(), newAlarm)
}
