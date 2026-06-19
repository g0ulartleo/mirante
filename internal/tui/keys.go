package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		m.cancel()
		return m, tea.Quit
	}

	if m.input != inputNone {
		return m.handleInputKey(msg)
	}

	if msg.String() == "q" {
		m.cancel()
		return m, tea.Quit
	}

	if m.showHelp {
		m.showHelp = false
		return m, nil
	}

	if m.mode == detailView {
		return m.handleDetailKey(msg)
	}
	return m.handleListKey(msg)
}

func (m *Model) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.input = inputNone
		m.inputBuf = ""
		return m, nil
	case "enter":
		mode := m.input
		buf := strings.TrimSpace(m.inputBuf)
		m.input = inputNone
		m.inputBuf = ""
		if mode == inputFilter {
			m.filter = buf
			m.alarmCursor = 0
			m.topRow = 0
			m.rebuildRows()
			m.ensureVisible()
			return m, nil
		}
		return m.runCommand(buf)
	case "backspace":
		if len(m.inputBuf) > 0 {
			r := []rune(m.inputBuf)
			m.inputBuf = string(r[:len(r)-1])
		}
		return m, nil
	case "ctrl+w", "alt+backspace":
		m.inputBuf = deleteLastInputWord(m.inputBuf)
		return m, nil
	default:
		if len(msg.Runes) > 0 {
			m.inputBuf += string(msg.Runes)
		}
		return m, nil
	}
}

func (m *Model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		m.cancel()
		return m, tea.Quit
	case "j", "down":
		if m.alarmCursor < len(m.alarmRows)-1 {
			m.alarmCursor++
			m.ensureVisible()
		}
	case "k", "up":
		if m.alarmCursor > 0 {
			m.alarmCursor--
			m.ensureVisible()
		}
	case "pgdown", "pagedown":
		if len(m.alarmRows) > 0 {
			m.alarmCursor += m.listBodyHeight()
			if m.alarmCursor >= len(m.alarmRows) {
				m.alarmCursor = len(m.alarmRows) - 1
			}
			m.ensureVisible()
		}
	case "pgup", "pageup":
		if len(m.alarmRows) > 0 {
			m.alarmCursor -= m.listBodyHeight()
			if m.alarmCursor < 0 {
				m.alarmCursor = 0
			}
			m.ensureVisible()
		}
	case "g", "home":
		m.alarmCursor = 0
		m.ensureVisible()
	case "G", "end":
		m.alarmCursor = len(m.alarmRows) - 1
		m.ensureVisible()
	case "l", "right", "enter":
		return m.openDetail()
	case "h", "left", "esc":
		if m.filter != "" {
			m.filter = ""
			m.alarmCursor = 0
			m.topRow = 0
			m.rebuildRows()
			m.ensureVisible()
		}
	case "/":
		m.input = inputFilter
		m.inputBuf = m.filter
		return m, nil
	case ":":
		m.input = inputCommand
		m.inputBuf = ""
		return m, nil
	case "s", "S":
		m.toggleSort()
		m.rebuildRows()
		m.ensureVisible()
	case "?":
		m.showHelp = true
	case "r":
		m.statusMsg = "refreshing…"
		return m, fetchCmd(m.client)
	case "tab":
		switch m.listMode {
		case triageList:
			m.listMode = allList
		default:
			m.listMode = triageList
		}
		m.alarmCursor = 0
		m.topRow = 0
		m.rebuildRows()
		m.ensureVisible()
	}
	return m, nil
}

func (m *Model) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "h", "left", "esc", "backspace", "tab":
		m.mode = listView
		return m, nil
	case "?":
		m.showHelp = true
		return m, nil
	case "y":
		if as, ok := m.selectedAlarm(); ok && as.Alarm.ID != "" {
			return m, copyAlarmIDCmd(as.Alarm.ID)
		}
		return m, nil
	case "r":
		m.statusMsg = "refreshing…"
		return m, tea.Batch(fetchCmd(m.client), fetchDetailCmd(m.client, m.selectedID))
	case "j", "down":
		n := len(m.detailSignals)
		if n > 0 && m.detailSignalCursor < n-1 {
			m.detailSignalCursor++
			m.updateDetailView()
		}
		return m, nil
	case "k", "up":
		if m.detailSignalCursor > 0 {
			m.detailSignalCursor--
			m.updateDetailView()
		}
		return m, nil
	case "]", "pgdown", "pagedown":
		n := len(m.detailSignals)
		if n > 0 {
			m.detailSignalCursor += detailSignalsPageSize
			if m.detailSignalCursor >= n {
				m.detailSignalCursor = n - 1
			}
			m.updateDetailView()
		}
		return m, nil
	case "[", "pgup", "pageup":
		if len(m.detailSignals) > 0 {
			m.detailSignalCursor -= detailSignalsPageSize
			if m.detailSignalCursor < 0 {
				m.detailSignalCursor = 0
			}
			m.updateDetailView()
		}
		return m, nil
	case "c":
		if len(m.detailSignals) == 0 {
			return m, nil
		}
		if m.detailSignalCursor >= len(m.detailSignals) {
			m.detailSignalCursor = len(m.detailSignals) - 1
		}
		sig := m.detailSignals[m.detailSignalCursor]
		return m, copySignalCmd(sig)
	}
	var cmd tea.Cmd
	m.detail, cmd = m.detail.Update(msg)
	return m, cmd
}

func deleteLastInputWord(s string) string {
	r := []rune(s)
	i := len(r)
	for i > 0 && r[i-1] == ' ' {
		i--
	}
	for i > 0 && r[i-1] != ' ' {
		i--
	}
	return string(r[:i])
}
