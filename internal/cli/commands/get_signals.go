package commands

import (
	"fmt"
	"time"

	"github.com/g0ulartleo/mirante/internal/cli"
	"github.com/g0ulartleo/mirante/internal/config"
)

type GetSignalsCommand struct{}

func (c *GetSignalsCommand) Name() string {
	return "get signals"
}

func (c *GetSignalsCommand) Description() string {
	return "Get the latest signals/status history for a specific alarm"
}

func (c *GetSignalsCommand) Usage() string {
	return "get signals <alarm-id>"
}

func (c *GetSignalsCommand) Run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: ./cli get signals <alarm-id>")
	}

	alarmID := args[0]

	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	apiClient := NewAPIClient(cliConfig)

	signals, err := apiClient.GetAlarmSignals(alarmID)
	if err != nil {
		return fmt.Errorf("failed to get alarm signals: %w", err)

	}
	for _, signal := range signals {
		fmt.Printf("[%s][%s]: %s\n", signal.Timestamp.Format(time.RFC3339), signal.Status, signal.Message)
	}
	return nil
}

func init() {
	c := &GetSignalsCommand{}
	cli.RegisterCommand(c.Name(), c)
	cli.RegisterAlias("get-signals", c)
}
