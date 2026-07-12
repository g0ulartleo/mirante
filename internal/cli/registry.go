package cli

import (
	"fmt"
	"strings"
)

var Version = "dev"

type Command interface {
	Name() string
	Description() string
	Usage() string
	Run(args []string) error
}

type CommandsRegistry map[string]Command

var Registry = &CommandsRegistry{}

func RegisterCommand(name string, command Command) {
	(*Registry)[name] = command
}

func RegisterAlias(alias string, command Command) {
	RegisterCommand(alias, command)
}

func GetCommand(name string) (Command, error) {
	command, exists := (*Registry)[name]
	if !exists {
		return nil, fmt.Errorf("command %s not found", name)
	}
	return command, nil
}

func ResolveCommand(args []string) (Command, []string, error) {
	for consumed := len(args); consumed > 0; consumed-- {
		name := strings.Join(args[:consumed], " ")
		command, exists := (*Registry)[name]
		if exists {
			return command, args[consumed:], nil
		}
	}
	if len(args) == 0 {
		return nil, nil, fmt.Errorf("command not found")
	}
	return nil, nil, fmt.Errorf("command %s not found", args[0])
}
