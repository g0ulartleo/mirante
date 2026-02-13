package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/g0ulartleo/mirante/internal/cli"
)

type HelpCommand struct{}

func (c *HelpCommand) Name() string {
	return "help"
}

func (c *HelpCommand) Description() string {
	return "Display help information about available commands"
}

func (c *HelpCommand) Usage() string {
	return "help [command]"
}

func (c *HelpCommand) Run(args []string) error {
	if len(args) > 0 {
		commandName := args[0]
		command, err := cli.GetCommand(commandName)
		if err != nil {
			return fmt.Errorf("command '%s' not found", commandName)
		}

		fmt.Printf("Command: %s\n", command.Name())
		fmt.Printf("Description: %s\n", command.Description())
		fmt.Printf("Usage: %s\n", command.Usage())
		return nil
	}

	fmt.Println("Mirante CLI - Monitor and manage your alarms")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  mirante <command> [arguments]")
	fmt.Println()
	fmt.Println("Available Commands:")

	commands := make([]cli.Command, 0, len(*cli.Registry))
	for _, command := range *cli.Registry {
		commands = append(commands, command)
	}

	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name() < commands[j].Name()
	})

	maxNameLength := 0
	for _, command := range commands {
		if len(command.Name()) > maxNameLength {
			maxNameLength = len(command.Name())
		}
	}
	for _, command := range commands {
		padding := strings.Repeat(" ", maxNameLength-len(command.Name())+2)
		fmt.Printf("  %s%s%s\n", command.Name(), padding, command.Description())
	}

	fmt.Println()
	fmt.Println("Use 'mirante help <command>' for more information about a command.")

	return nil
}

func init() {
	c := &HelpCommand{}
	cli.RegisterCommand(c.Name(), c)
}
