package commands

import (
	"fmt"
	"log"

	"github.com/g0ulartleo/mirante/internal/cli"
	"github.com/g0ulartleo/mirante/internal/config"
)

type RunAlarmCommand struct{}

func (c *RunAlarmCommand) Name() string {
	return "run alarm"
}

func (c *RunAlarmCommand) Description() string {
	return "Manually trigger a health check for a specific alarm"
}

func (c *RunAlarmCommand) Usage() string {
	return "run alarm <alarm-id>"
}

func (c *RunAlarmCommand) Run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: ./cli %s <alarm-id>", c.Name())
	}

	alarmID := args[0]
	config, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load CLI config: %v", err)
	}
	client := NewAPIClient(config)
	if err := client.RunAlarm(alarmID); err != nil {
		return fmt.Errorf("failed to run alarm: %v", err)
	}

	log.Printf("Task enqueued")
	return nil
}

func init() {
	c := &RunAlarmCommand{}
	cli.RegisterCommand(c.Name(), c)
	cli.RegisterAlias("run-alarm", c)
}
