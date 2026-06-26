package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type commandAction int

const (
	commandRunAlarm commandAction = iota
	commandRefresh
	commandSyncAlarms
)

type commandPaletteItem struct {
	label  string
	help   string
	action commandAction
}

var commandPaletteItems = []commandPaletteItem{
	{label: "run alarm", help: "selected/current alarm", action: commandRunAlarm},
	{label: "refresh", help: "fetch latest alarms", action: commandRefresh},
	{label: "sync alarms", help: "sync runtime alarms", action: commandSyncAlarms},
}

func (m *Model) openCommandPalette() {
	m.input = inputNone
	m.inputBuf = ""
	m.leaderMode = false
	m.commandPaletteOpen = true
	if m.commandPaletteCursor >= len(commandPaletteItems) {
		m.commandPaletteCursor = 0
	}
}

func (m *Model) handleCommandPaletteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+p", "esc":
		m.commandPaletteOpen = false
		return m, nil
	case "j", "down":
		if m.commandPaletteCursor < len(commandPaletteItems)-1 {
			m.commandPaletteCursor++
		}
		return m, nil
	case "k", "up":
		if m.commandPaletteCursor > 0 {
			m.commandPaletteCursor--
		}
		return m, nil
	case "enter":
		m.commandPaletteOpen = false
		return m.executeCommandAction(commandPaletteItems[m.commandPaletteCursor].action)
	}
	return m, nil
}

func (m *Model) overlayCommandPalette() string {
	rows := []string{panelTitleStyle.Render("COMMANDS"), ""}
	for i, item := range commandPaletteItems {
		marker := "  "
		fg := colorText
		bg := lipgloss.Color("")
		if i == m.commandPaletteCursor {
			marker = "▶ "
			fg = colorBright
			bg = colorSelBg
		}
		line := lipgloss.NewStyle().Foreground(fg).Background(bg).Render(marker + pad(item.label, 14, false) + " " + item.help)
		rows = append(rows, line)
	}
	rows = append(rows, "", keyRow("ctrl+x r", "run alarm", 42), keyRow("ctrl+x s", "sync alarms", 42))
	box := panelBoxStyle.Width(48).BorderForeground(colorAccent).Render(strings.Join(rows, "\n"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m *Model) executeCommandAction(action commandAction) (tea.Model, tea.Cmd) {
	switch action {
	case commandRunAlarm:
		return m.runSelectedAlarm()
	case commandRefresh:
		return m.refreshCurrentView()
	case commandSyncAlarms:
		m.statusMsg = "syncing…"
		return m, syncAlarmsCmd(m.client)
	default:
		return m, nil
	}
}

func (m *Model) runSelectedAlarm() (tea.Model, tea.Cmd) {
	id := m.currentAlarmID()
	if id == "" {
		m.statusMsg = "no alarm selected"
		return m, nil
	}
	m.statusMsg = "running…"
	return m, runAlarmCmd(m.client, id)
}

func (m *Model) refreshCurrentView() (tea.Model, tea.Cmd) {
	m.statusMsg = "refreshing…"
	if m.mode == detailView && m.selectedID != "" {
		return m, tea.Batch(fetchCmd(m.client), fetchDetailCmd(m.client, m.selectedID))
	}
	return m, fetchCmd(m.client)
}

func (m *Model) currentAlarmID() string {
	if m.mode == detailView && m.selectedID != "" {
		return m.selectedID
	}
	if as, ok := m.selectedRowAlarm(); ok {
		return as.Alarm.ID
	}
	return ""
}
