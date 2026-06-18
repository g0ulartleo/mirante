package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCommand struct {
	name string
}

func (c fakeCommand) Name() string        { return c.name }
func (c fakeCommand) Description() string { return "test" }
func (c fakeCommand) Usage() string       { return c.name }
func (c fakeCommand) Run(args []string) error {
	return nil
}

func TestResolveCommandUsesLongestPrefix(t *testing.T) {
	original := Registry
	testRegistry := CommandsRegistry{}
	Registry = &testRegistry
	t.Cleanup(func() { Registry = original })

	RegisterCommand("get", fakeCommand{name: "get"})
	RegisterCommand("get alarm", fakeCommand{name: "get alarm"})

	command, args, err := ResolveCommand([]string{"get", "alarm", "alarm-1"})
	require.NoError(t, err)
	assert.Equal(t, "get alarm", command.Name())
	assert.Equal(t, []string{"alarm-1"}, args)
}

func TestResolveCommandReturnsMissingCommand(t *testing.T) {
	original := Registry
	testRegistry := CommandsRegistry{}
	Registry = &testRegistry
	t.Cleanup(func() { Registry = original })

	_, _, err := ResolveCommand([]string{"plugin", "init"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "command plugin not found")
}
