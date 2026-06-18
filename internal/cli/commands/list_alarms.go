package commands

import (
	"fmt"
	"slices"
	"strings"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/cli"
	"github.com/g0ulartleo/mirante/internal/config"
)

type ListAlarmsCommand struct{}

func (c *ListAlarmsCommand) Name() string {
	return "list alarms"
}

func (c *ListAlarmsCommand) Description() string {
	return "List all configured alarms"
}

func (c *ListAlarmsCommand) Usage() string {
	return "list alarms"
}

func (c *ListAlarmsCommand) Run(args []string) error {
	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	apiClient := NewAPIClient(cliConfig)
	alarms, err := apiClient.ListAlarms()
	if err != nil {
		return fmt.Errorf("failed to list alarms: %w", err)
	}

	if len(alarms) == 0 {
		fmt.Println("No alarms found.")
		return nil
	}

	pathTree := make(map[string]interface{})

	for _, alarmItem := range alarms {
		current := pathTree

		for _, pathSegment := range alarmItem.Path {
			if current[pathSegment] == nil {
				current[pathSegment] = make(map[string]interface{})
			}
			current = current[pathSegment].(map[string]interface{})
		}

		if current["_alarms"] == nil {
			current["_alarms"] = []alarm.Alarm{}
		}
		alarmsSlice := current["_alarms"].([]alarm.Alarm)
		current["_alarms"] = append(alarmsSlice, alarmItem)
	}

	fmt.Printf("Found %d alarm(s):\n\n", len(alarms))
	printPathTree(pathTree, 0)

	return nil
}

func printPathTree(tree map[string]interface{}, level int) {
	var pathKeys []string
	for key := range tree {
		if key != "_alarms" {
			pathKeys = append(pathKeys, key)
		}
	}
	slices.Sort(pathKeys)

	indent := strings.Repeat("  ", level)

	for _, key := range pathKeys {
		fmt.Printf("%s▸ %s/\n", indent, key)
		printPathTree(tree[key].(map[string]interface{}), level+1)
		if level == 0 {
			fmt.Println()
		}
	}

	if alarmsInterface, exists := tree["_alarms"]; exists {
		alarms := alarmsInterface.([]alarm.Alarm)

		slices.SortFunc(alarms, func(a, b alarm.Alarm) int {
			return strings.Compare(a.Name, b.Name)
		})

		for i, a := range alarms {
			fmt.Printf("%s• %s\n", indent, a.Name)
			fmt.Printf("%s  ID: \033[1m%s\033[0m\n", indent, a.ID)
			fmt.Printf("%s  Description: %s\n", indent, a.Description)

			if i < len(alarms)-1 {
				fmt.Println()
			}
		}
	}
}

func init() {
	c := &ListAlarmsCommand{}
	cli.RegisterCommand(c.Name(), c)
	cli.RegisterAlias("list-alarms", c)
}
