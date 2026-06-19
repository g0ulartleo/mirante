package commands

import (
	"github.com/g0ulartleo/mirante/internal/apiclient"
	"github.com/g0ulartleo/mirante/internal/config"
)

// Client is an alias to the shared API client so existing CLI commands keep
// working unchanged while the TUI/SSH app reuse the same implementation.
type Client = apiclient.Client

func NewAPIClient(config *config.CLIConfig) *Client {
	return apiclient.New(config)
}
