package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) overlayHelp() string {
	help := []string{
		panelTitleStyle.Render("HELP"),
		"",
		keyRow("j / k", "move cursor", 30),
		keyRow("[ / ]", "page signals", 30),
		keyRow("l / → / ↵", "open alarm", 30),
		keyRow("h / ← / esc", "back · clear filter", 30),
		keyRow("/", "filter", 30),
		keyRow(":", "command", 30),
		keyRow("s", "toggle sort", 30),
		keyRow("tab", "toggle triage/all view", 30),
		keyRow("r", "refresh", 30),
		keyRow("y", "copy alarm id", 30),
		keyRow("c", "copy signal", 30),
		keyRow("enter", "expand signal", 30),
		keyRow("q", "quit", 30),
	}
	box := panelBoxStyle.Width(34).BorderForeground(colorAccent).Render(strings.Join(help, "\n"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
