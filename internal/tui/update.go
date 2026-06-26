package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/g0ulartleo/mirante/internal/alarm"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case marqueeTickMsg:
		m.marqueeOffset++
		return m, marqueeTickCmd()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeDetail()
		m.ensureVisible()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case signalsMsg:
		m.applyData([]alarm.AlarmSignals(msg))
		return m, nil

	case wsUpdateMsg:
		m.applyData([]alarm.AlarmSignals(msg))
		return m, listenCmd(m.ctx, m.updates, m.errs)

	case connectedMsg:
		m.updates = msg.updates
		m.errs = msg.errs
		m.connected = true
		m.err = nil
		return m, listenCmd(m.ctx, m.updates, m.errs)

	case wsClosedMsg:
		m.connected = false
		if m.ctx.Err() != nil {
			return m, nil
		}
		return m, reconnectCmd(m.ctx, m.client)

	case detailSignalsMsg:
		if msg.id == m.selectedID {
			m.detailSignals = normalizeDetailSignals(msg.signals, detailSignalsFetchLimit)
			m.detailHistoryAt = newestSignalTime(m.detailSignals)
			if m.detailHistoryAt.IsZero() {
				m.detailHistoryAt = time.Now()
			}
			m.detailLoading = false
			m.detailSignalCursor = 0
			if as, ok := m.selectedAlarm(); ok {
				m.detail.SetContent(m.renderDetail(as))
			}
		}
		return m, nil

	case errMsg:
		m.err = msg.err
		m.connected = false
		return m, nil

	case successMsg:
		m.statusMsg = msg.text
		return m, nil
	}

	if m.mode == detailView {
		var cmd tea.Cmd
		m.detail, cmd = m.detail.Update(msg)
		return m, cmd
	}
	return m, nil
}
