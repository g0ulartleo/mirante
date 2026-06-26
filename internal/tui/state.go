package tui

import (
	"time"

	"github.com/g0ulartleo/mirante/internal/alarm"
)

func (m *Model) rebuildRows() {
	data := m.all
	if m.listMode == triageList {
		var filtered []alarm.AlarmSignals
		for _, a := range data {
			if severityRank(lastStatus(a)) >= 1 {
				filtered = append(filtered, a)
			}
		}
		data = filtered
	}
	m.rows, m.alarmRows = buildRows(data, m.filter, m.sort)
	if m.alarmCursor >= len(m.alarmRows) {
		m.alarmCursor = len(m.alarmRows) - 1
	}
	if m.alarmCursor < 0 {
		m.alarmCursor = 0
	}
}

func (m *Model) activeRowIndex() int {
	if m.alarmCursor < 0 || m.alarmCursor >= len(m.alarmRows) {
		return -1
	}
	return m.alarmRows[m.alarmCursor]
}

func (m *Model) selectedRowAlarm() (alarm.AlarmSignals, bool) {
	ri := m.activeRowIndex()
	if ri < 0 {
		return alarm.AlarmSignals{}, false
	}
	return m.rows[ri].alarm, true
}

func (m *Model) selectedAlarm() (alarm.AlarmSignals, bool) {
	for _, a := range m.all {
		if a.Alarm.ID == m.selectedID {
			return a, true
		}
	}
	return alarm.AlarmSignals{}, false
}

func (m *Model) overallHealth() healthCounts {
	return countHealth(m.all)
}

func (m *Model) applyData(data []alarm.AlarmSignals) {
	m.all = data
	m.lastUpdate = time.Now()
	m.statusMsg = ""
	m.rebuildRows()
	m.ensureVisible()
	if m.mode == detailView {
		if as, ok := m.selectedAlarm(); ok {
			m.detailSignals = mergeDetailSignals(m.detailSignals, as.Signals, detailSignalsFetchLimit)
			m.detailHistoryAt = newestSignalTime(m.detailSignals)
			if m.detailHistoryAt.IsZero() {
				m.detailHistoryAt = time.Now()
			}
			m.detail.SetContent(m.renderDetail(as))
		} else {
			m.mode = listView
		}
	}
}

func (m *Model) ensureVisible() {
	ri := m.activeRowIndex()
	if ri < 0 {
		m.topRow = 0
		return
	}
	h := m.listBodyHeight()
	if h < 1 {
		return
	}
	if ri < m.topRow {
		m.topRow = ri
	}
	if ri >= m.topRow+h {
		m.topRow = ri - h + 1
	}
	if ri > 0 && m.rows[ri].kind == rowAlarm && m.topRow == ri {
		if m.rows[ri-1].kind == rowGroupHeader {
			m.topRow = ri - 1
		}
	}
	if m.topRow < 0 {
		m.topRow = 0
	}
}
