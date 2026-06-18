package commands

import (
	"testing"

	"github.com/g0ulartleo/mirante/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupedReadCheckCommandsAreRegistered(t *testing.T) {
	tests := []struct {
		args []string
		name string
	}{
		{args: []string{"list", "alarms"}, name: "list alarms"},
		{args: []string{"get", "alarm", "alarm-1"}, name: "get alarm"},
		{args: []string{"get", "signals", "alarm-1"}, name: "get signals"},
		{args: []string{"run", "alarm", "alarm-1"}, name: "run alarm"},
		{args: []string{"init", "repo", "--runtime", "nodejs", "--dir", "runtime"}, name: "init repo"},
		{args: []string{"new", "alarm", "alarm-1"}, name: "new alarm"},
	}

	for _, tt := range tests {
		command, _, err := cli.ResolveCommand(tt.args)
		require.NoError(t, err)
		assert.Equal(t, tt.name, command.Name())
	}
}

func TestFlatReadCheckAliasesStillResolve(t *testing.T) {
	tests := []struct {
		args []string
		name string
	}{
		{args: []string{"list-alarms"}, name: "list alarms"},
		{args: []string{"get-alarm", "alarm-1"}, name: "get alarm"},
		{args: []string{"get-signals", "alarm-1"}, name: "get signals"},
		{args: []string{"run-alarm", "alarm-1"}, name: "run alarm"},
	}

	for _, tt := range tests {
		command, _, err := cli.ResolveCommand(tt.args)
		require.NoError(t, err)
		assert.Equal(t, tt.name, command.Name())
	}
}

func TestLegacyWriteAndPluginCommandsAreUnavailable(t *testing.T) {
	missing := [][]string{
		{"set-alarm", "alarm.yml"},
		{"delete-alarm", "alarm-1"},
		{"plugin", "init"},
		{"plugin", "validate"},
		{"plugin", "check"},
	}

	for _, args := range missing {
		_, _, err := cli.ResolveCommand(args)
		require.Error(t, err)
	}
}
