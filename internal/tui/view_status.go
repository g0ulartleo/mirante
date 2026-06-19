package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) statusBar() string {
	mode := "NORMAL"
	modeColor := colorAccent
	if m.filter != "" {
		mode = "FILTER"
		modeColor = colorWarning
	}
	switch m.input {
	case inputFilter:
		mode = "FILTER"
		modeColor = colorWarning
	case inputCommand:
		mode = "COMMAND"
		modeColor = colorHealthy
	}
	badge := lipgloss.NewStyle().Foreground(colorBg).Background(modeColor).Bold(true).Padding(0, 1).Render(mode)
	var left string
	if m.input != inputNone {
		prompt := "/"
		if m.input == inputCommand {
			prompt = ":"
		}
		left = badge + " " + lipgloss.NewStyle().Foreground(colorBright).Render(prompt+m.inputBuf+"█")
	} else {
		parts := []string{"mirante", m.listMode.String()}
		if m.filter != "" {
			parts = append(parts, m.filter)
		}
		crumb := breadcrumbStyle.Render(fmt.Sprintf("%s [%d/%d]", strings.Join(parts, " › "), m.alarmCursor+1, len(m.alarmRows)))
		if m.statusMsg != "" {
			crumb = breadcrumbStyle.Render(m.statusMsg)
		}
		left = badge + " " + crumb
	}
	status := "● live"
	statusColor := colorHealthy
	if !m.connected {
		status = "○ offline"
		statusColor = colorUnhealthy
	}
	right := lipgloss.NewStyle().Foreground(statusColor).Render(status)
	if !m.lastUpdate.IsZero() {
		right += breadcrumbStyle.Render(" · refreshed " + ageString(m.lastUpdate))
	}
	right += breadcrumbStyle.Render(" · :")
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}
