package commands

import (
	"fmt"

	"github.com/g0ulartleo/mirante/internal/cli"
)

type VersionCommand struct{}

func (c *VersionCommand) Name() string {
	return "version"
}

func (c *VersionCommand) Description() string {
	return "Print the version number"
}

func (c *VersionCommand) Usage() string {
	return "version"
}

func (c *VersionCommand) Run(args []string) error {
	fmt.Println(cli.Version)
	return nil
}

func init() {
	c := &VersionCommand{}
	cli.RegisterCommand(c.Name(), c)
}
