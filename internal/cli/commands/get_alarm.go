package commands

import (
	"fmt"

	"github.com/g0ulartleo/mirante/internal/cli"
	"github.com/g0ulartleo/mirante/internal/config"
	"gopkg.in/yaml.v3"
)

type GetAlarmCommand struct{}

func (c *GetAlarmCommand) Name() string {
	return "get-alarm"
}

func (c *GetAlarmCommand) Description() string {
	return "Get detailed information about a specific alarm in YAML format"
}

func (c *GetAlarmCommand) Usage() string {
	return "get-alarm <alarm-id>"
}

func (c *GetAlarmCommand) Run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: ./cli get-alarm <alarm-id>")
	}

	alarmID := args[0]

	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	apiClient := NewAPIClient(cliConfig)

	alarm, err := apiClient.GetAlarm(alarmID)
	if err != nil {
		return fmt.Errorf("failed to get alarm: %w", err)
	}
	yamlData, err := yaml.Marshal(alarm)
	if err != nil {
		return fmt.Errorf("failed to marshal alarm: %w", err)
	}
	fmt.Printf("%s", yamlData)
	return nil
}

func init() {
	c := &GetAlarmCommand{}
	cli.RegisterCommand(c.Name(), c)
}
