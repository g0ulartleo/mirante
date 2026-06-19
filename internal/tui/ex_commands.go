package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) runCommand(cmd string) (tea.Model, tea.Cmd) {
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return m, nil
	}
	switch fields[0] {
	case "q", "quit", "q!":
		m.cancel()
		return m, tea.Quit
	case "sync":
		m.statusMsg = "syncing…"
		return m, syncAlarmsCmd(m.client)
	case "run":
		id := m.selectedID
		if id == "" {
			if as, ok := m.selectedRowAlarm(); ok {
				id = as.Alarm.ID
			}
		}
		if id == "" {
			m.statusMsg = "no alarm selected"
			return m, nil
		}
		m.statusMsg = "running…"
		return m, runAlarmCmd(m.client, id)
	case "refresh", "r":
		m.statusMsg = "refreshing…"
		return m, fetchCmd(m.client)
	case "sort":
		if len(fields) > 1 {
			switch fields[1] {
			case "name":
				m.sort = sortName
			case "severity", "sev":
				m.sort = sortSeverity
			}
		} else {
			m.toggleSort()
		}
		m.rebuildRows()
		m.ensureVisible()
		return m, nil
	case "filter":
		m.filter = strings.TrimSpace(strings.TrimPrefix(cmd, "filter"))
		m.alarmCursor = 0
		m.topRow = 0
		m.rebuildRows()
		m.ensureVisible()
		return m, nil
	case "clear":
		m.filter = ""
		m.rebuildRows()
		m.ensureVisible()
		return m, nil
	default:
		m.statusMsg = "unknown command: " + fields[0]
		return m, nil
	}
}

func (m *Model) toggleSort() {
	if m.sort == sortSeverity {
		m.sort = sortName
	} else {
		m.sort = sortSeverity
	}
}
