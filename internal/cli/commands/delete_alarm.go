package commands

import (
	"fmt"
	"log"

	"github.com/g0ulartleo/mirante/internal/cli"
	"github.com/g0ulartleo/mirante/internal/config"
)

type DeleteAlarmCommand struct{}

func (c *DeleteAlarmCommand) Name() string {
	return "delete-alarm"
}

func (c *DeleteAlarmCommand) Description() string {
	return "Delete an alarm by its ID"
}

func (c *DeleteAlarmCommand) Usage() string {
	return "delete-alarm <alarm-id>"
}

func (c *DeleteAlarmCommand) Run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: ./cli %s <alarm-id>", c.Name())
	}
	alarmID := args[0]
	config, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load CLI config: %v", err)
	}
	client := NewAPIClient(config)
	if err := client.DeleteAlarm(alarmID); err != nil {
		return fmt.Errorf("failed to delete alarm: %v", err)
	}
	log.Printf("Alarm %s deleted", alarmID)
	return nil
}

func init() {
	c := &DeleteAlarmCommand{}
	cli.RegisterCommand(c.Name(), c)
}
