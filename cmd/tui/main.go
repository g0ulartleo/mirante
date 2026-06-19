package main

import (
	"fmt"
	"os"

	"github.com/g0ulartleo/mirante/internal/apiclient"
	"github.com/g0ulartleo/mirante/internal/config"
	"github.com/g0ulartleo/mirante/internal/tui"
)

func main() {
	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	client := apiclient.New(cliConfig)
	if err := tui.Run(client); err != nil {
		fmt.Fprintf(os.Stderr, "tui error: %v\n", err)
		os.Exit(1)
	}
}
