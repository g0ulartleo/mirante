package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/g0ulartleo/mirante/internal/signal"
)

func (m *Model) openDetail() (tea.Model, tea.Cmd) {
	as, ok := m.selectedRowAlarm()
	if !ok {
		return m, nil
	}
	m.selectedID = as.Alarm.ID
	m.mode = detailView
	m.detailSignals = nil
	m.detailHistoryAt = time.Time{}
	m.detailLoading = true
	m.detailSignalCursor = 0
	m.resizeDetail()
	m.detail.SetContent(m.renderDetail(as))
	m.detail.GotoTop()
	return m, fetchDetailCmd(m.client, m.selectedID)
}

func (m *Model) resizeDetail() {
	if m.width <= 0 || m.height <= 0 {
		return
	}
	h := m.height - 2
	if h < 1 {
		h = 1
	}
	if !m.detailReady {
		m.detail = viewport.New(m.width, h)
		m.detailReady = true
	} else {
		m.detail.Width = m.width
		m.detail.Height = h
	}
}

func (m *Model) updateDetailView() {
	as, ok := m.selectedAlarm()
	if !ok {
		return
	}
	y := m.detail.YOffset
	content := m.renderDetail(as)
	m.detail.SetContent(content)
	contentH := lipgloss.Height(content)
	vpH := m.detail.Height
	if y > contentH-vpH {
		y = contentH - vpH
	}
	if y < 0 {
		y = 0
	}
	m.detail.SetYOffset(y)
}

func newestSignalTime(signals []signal.Signal) time.Time {
	var newest time.Time
	for _, s := range signals {
		if s.Timestamp.After(newest) {
			newest = s.Timestamp
		}
	}
	return newest
}
