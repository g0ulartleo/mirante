package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/g0ulartleo/mirante/internal/apiclient"
)

func Run(client *apiclient.Client) error {
	p := tea.NewProgram(NewModel(client), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
