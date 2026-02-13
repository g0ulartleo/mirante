package commands

import (
	"fmt"
	"strings"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/cli"
	"github.com/g0ulartleo/mirante/internal/config"
)

type SetAlarmCommand struct{}

func (c *SetAlarmCommand) Name() string {
	return "set-alarm"
}

func (c *SetAlarmCommand) Description() string {
	return "Create or update an alarm from a YAML configuration file"
}

func (c *SetAlarmCommand) Usage() string {
	return "set-alarm <file>"
}

func (c *SetAlarmCommand) Run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: ./cli %s <file>", c.Name())
	}
	filePath := args[0]
	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var a *alarm.Alarm
	if strings.HasSuffix(filePath, ".yaml") || strings.HasSuffix(filePath, ".yml") {
		a, err = alarm.LoadAlarmConfig(filePath)
		if err != nil {
			return fmt.Errorf("failed to load alarm: %w", err)
		}
	} else if strings.HasSuffix(filePath, ".json") {
		return fmt.Errorf("json files are not supported yet")
	} else {
		return fmt.Errorf("invalid file type: %s", filePath)
	}

	apiClient := NewAPIClient(cliConfig)
	err = apiClient.SetAlarm(a)
	if err != nil {
		return fmt.Errorf("failed to create or update alarm: %w", err)
	}
	return nil
}

func init() {
	c := &SetAlarmCommand{}
	cli.RegisterCommand(c.Name(), c)
}
