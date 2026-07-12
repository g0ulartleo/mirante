package commands

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/g0ulartleo/mirante/internal/cli"
	"gopkg.in/yaml.v3"
)

//go:embed all:scaffold
var scaffoldFS embed.FS

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
		return scaffoldRepo(dir, "scaffold/nodejs")
	case "go":
		return scaffoldRepo(dir, "scaffold/go")
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
		content, err := renderAlarmTemplate("scaffold/templates/node_alarm.ts.tmpl", nodeAlarmData(alarmID))
		if err != nil {
			return err
		}
		return writeFileExclusive(path, content)
	case "go":
		path := filepath.Join(marker.AlarmsDir, strings.ReplaceAll(alarmID, "-", "_")+".go")
		content, err := renderAlarmTemplate("scaffold/templates/go_alarm.go.tmpl", goAlarmData(alarmID))
		if err != nil {
			return err
		}
		return writeFileExclusive(path, content)
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

func scaffoldRepo(dir string, root string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create runtime directory: %w", err)
	}

	return fs.WalkDir(scaffoldFS, root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		rel := strings.TrimSuffix(strings.TrimPrefix(path, root+"/"), ".tmpl")
		content, err := scaffoldFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read scaffold file %s: %w", path, err)
		}
		return writeFileExclusive(filepath.Join(dir, filepath.FromSlash(rel)), string(content))
	})
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

type alarmTemplateData struct {
	AlarmID    string
	ExportName string
	GoName     string
	HumanName  string
}

func nodeAlarmData(alarmID string) alarmTemplateData {
	className := toPascalCase(alarmID)
	return alarmTemplateData{
		AlarmID:    alarmID,
		ExportName: strings.ToLower(className[:1]) + className[1:],
		HumanName:  humanizeAlarmID(alarmID),
	}
}

func goAlarmData(alarmID string) alarmTemplateData {
	return alarmTemplateData{
		AlarmID:   alarmID,
		GoName:    toPascalCase(alarmID),
		HumanName: humanizeAlarmID(alarmID),
	}
}

func renderAlarmTemplate(path string, data alarmTemplateData) (string, error) {
	content, err := scaffoldFS.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read alarm template %s: %w", path, err)
	}
	tmpl, err := template.New(filepath.Base(path)).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse alarm template %s: %w", path, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render alarm template %s: %w", path, err)
	}
	return buf.String(), nil
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
