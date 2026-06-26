package tui

import (
	"sort"
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

func normalizeDetailSignals(signals []signal.Signal, limit int) []signal.Signal {
	return mergeDetailSignals(nil, signals, limit)
}

func mergeDetailSignals(existing, incoming []signal.Signal, limit int) []signal.Signal {
	if len(existing) == 0 && len(incoming) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(existing)+len(incoming))
	merged := make([]signal.Signal, 0, len(existing)+len(incoming))
	add := func(signals []signal.Signal) {
		for _, s := range signals {
			key := detailSignalKey(s)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			merged = append(merged, s)
		}
	}
	add(existing)
	add(incoming)
	sort.SliceStable(merged, func(i, j int) bool {
		return merged[i].Timestamp.After(merged[j].Timestamp)
	})
	if limit > 0 && len(merged) > limit {
		merged = merged[:limit]
	}
	return merged
}

func detailSignalKey(s signal.Signal) string {
	return s.AlarmID + "\x00" + s.Timestamp.UTC().Format(time.RFC3339Nano) + "\x00" + string(s.Status) + "\x00" + s.Message
}
