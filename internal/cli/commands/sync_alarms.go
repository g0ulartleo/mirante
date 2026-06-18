package commands

import (
	"fmt"
	"log"

	"github.com/g0ulartleo/mirante/internal/cli"
	"github.com/g0ulartleo/mirante/internal/config"
)

type SyncAlarmsCommand struct{}

func (c *SyncAlarmsCommand) Name() string {
	return "sync alarms"
}

func (c *SyncAlarmsCommand) Description() string {
	return "Sync alarms from all runtime repositories"
}

func (c *SyncAlarmsCommand) Usage() string {
	return "sync alarms"
}

func (c *SyncAlarmsCommand) Run(args []string) error {
	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	client := NewAPIClient(cliConfig)
	if err := client.SyncAlarms(); err != nil {
		return fmt.Errorf("failed to sync alarms: %w", err)
	}

	log.Printf("Alarms synced successfully")
	return nil
}

func init() {
	c := &SyncAlarmsCommand{}
	cli.RegisterCommand(c.Name(), c)
	cli.RegisterAlias("sync-alarms", c)
}
