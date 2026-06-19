package commands

import (
	"fmt"

	"github.com/g0ulartleo/mirante/internal/apiclient"
	"github.com/g0ulartleo/mirante/internal/cli"
	"github.com/g0ulartleo/mirante/internal/config"
	"github.com/g0ulartleo/mirante/internal/tui"
)

type TUICommand struct{}

func (c *TUICommand) Name() string {
	return "tui"
}

func (c *TUICommand) Description() string {
	return "Open the interactive terminal dashboard"
}

func (c *TUICommand) Usage() string {
	return "tui"
}

func (c *TUICommand) Run(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: mirante tui")
	}
	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	client := apiclient.New(cliConfig)
	if err := tui.Run(client); err != nil {
		return fmt.Errorf("tui error: %w", err)
	}
	return nil
}

func init() {
	c := &TUICommand{}
	cli.RegisterCommand(c.Name(), c)
}
