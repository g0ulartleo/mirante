package commands

import (
	"fmt"

	"github.com/g0ulartleo/mirante/internal/cli"
	"github.com/g0ulartleo/mirante/internal/config"
)

type AuthKeyCommand struct{}

func (c *AuthKeyCommand) Name() string {
	return "auth-key"
}

func (c *AuthKeyCommand) Description() string {
	return "Authenticate with the Mirante server using an API key"
}

func (c *AuthKeyCommand) Usage() string {
	return "auth-key <api_host> <api_key>"
}

func (c *AuthKeyCommand) Run(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: ./cli auth-key <api_host> <api_key>")
	}

	apiHost := args[0]
	apiKey := args[1]

	cliConfig := &config.CLIConfig{
		APIHost:   apiHost,
		APIKey:    apiKey,
		AuthType:  "api_key",
        AuthToken: "",
	}

	return config.SaveCLIConfig(cliConfig)
}

func init() {
	c := &AuthKeyCommand{}
	cli.RegisterCommand(c.Name(), c)
}
