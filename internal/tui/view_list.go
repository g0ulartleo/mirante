package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/g0ulartleo/mirante/internal/alarm"
)

const columnHeaderHeight = 1
const statusBarHeight = 1

type colLayout struct {
	marker int
	state  int
	alarm  int
	query  int
	last   int
	trend  int
}

func (m *Model) listScreen() string {
	header := m.headerBand()
	colHeader := m.columnHeader()
	status := m.statusBar()
	bodyHeight := m.listBodyHeight()
	var body string
	switch {
	case m.err != nil && len(m.all) == 0:
		body = errorStyle.Render("Error: " + m.err.Error())
	case m.listMode == triageList && m.filter == "" && len(m.all) > 0 && len(m.rows) == 0:
		body = m.peacefulTriage(bodyHeight)
	case len(m.rows) == 0:
		body = lipgloss.NewStyle().Padding(1, 2).Foreground(colorMuted).Render("No alarms match.")
	default:
		body = m.renderListBody(bodyHeight)
	}
	body = lipgloss.NewStyle().Height(bodyHeight).Render(body)
	return lipgloss.JoinVertical(lipgloss.Left, header, colHeader, body, status)
}

func (m *Model) peacefulTriage(height int) string {
	frames := []string{"☁", "☁☁", "☁ ☁", " ☁☁"}
	frame := frames[m.marqueeOffset%len(frames)]
	art := strings.Join([]string{
		lipgloss.NewStyle().Foreground(colorHealthy).Bold(true).Render("All peaceful for now"),
		lipgloss.NewStyle().Foreground(colorMuted).Render("No alarms need triage."),
		lipgloss.NewStyle().Foreground(colorAccent).Render("      " + frame),
		lipgloss.NewStyle().Foreground(colorFaint).Render("   .-^-._.-^-._.-^-."),
	}, "\n")
	return lipgloss.Place(m.width, height, lipgloss.Center, lipgloss.Center, art)
}

func (m *Model) listBodyHeight() int {
	h := m.height - lipgloss.Height(m.headerBand()) - columnHeaderHeight - statusBarHeight
	if h < 1 {
		h = 1
	}
	return h
}

func (m *Model) renderListBody(height int) string {
	end := m.topRow + height
	if end > len(m.rows) {
		end = len(m.rows)
	}
	var lines []string
	for i := m.topRow; i < end; i++ {
		r := m.rows[i]
		if r.kind == rowGroupHeader {
			lines = append(lines, m.renderGroupHeader(r))
		} else {
			selected := i == m.activeRowIndex()
			lines = append(lines, m.renderAlarmRow(r.alarm, selected))
		}
	}
	return strings.Join(lines, "\n")
}

func (m *Model) cols() colLayout {
	trend := 16
	switch {
	case m.width < 80:
		trend = 8
	case m.width < 100:
		trend = 12
	}
	c := colLayout{marker: 2, state: 10, last: 6, trend: trend}
	seps := 5
	fixed := c.marker + c.state + c.last + c.trend + seps
	remaining := m.width - fixed
	if remaining < 18 {
		remaining = 18
	}
	c.alarm = remaining * 70 / 100
	if c.alarm > 38 {
		c.alarm = 38
	}
	if c.alarm < 20 {
		c.alarm = 20
	}
	c.query = remaining - c.alarm
	if c.query < 4 {
		c.query = 4
	}
	return c
}

func (m *Model) columnHeader() string {
	c := m.cols()
	cells := []string{
		pad("", c.marker, false),
		pad("STATE", c.state, false),
		pad("ALARM", c.alarm, false),
		pad("SIGNAL", c.query, false),
		pad("LAST", c.last, true),
		pad("TREND", c.trend, false),
	}
	line := strings.Join(cells, " ")
	return colHeaderStyle.Render(line)
}

func (m *Model) renderGroupHeader(r listRow) string {
	name := strings.ToUpper(r.groupName)
	return fmt.Sprintf("▾ %s %s", statusSquare(groupColor(r.groupHealth)), groupHeaderStyle.Render(name))
}

func groupColor(h healthCounts) lipgloss.Color {
	switch {
	case h.unhealthy > 0:
		return colorUnhealthy
	case h.warning > 0:
		return colorWarning
	case h.healthy == h.total() && h.total() > 0:
		return colorHealthy
	default:
		return colorUnknown
	}
}

func (m *Model) renderAlarmRow(a alarm.AlarmSignals, selected bool) string {
	c := m.cols()
	st := lastStatus(a)
	marker := "  "
	if selected {
		marker = "▶ "
	}
	signalText := ""
	if len(a.Signals) > 0 {
		signalText = a.Signals[len(a.Signals)-1].Message
	}
	var bg lipgloss.Color
	if selected {
		bg = colorSelBg
	}
	signalCell := cellStyled(signalText, c.query, colorMuted, false, false, bg)
	if selected && lipgloss.Width(signalText) > c.query {
		signalCell = renderSignalCell(signalText, c.query, m.marqueeOffset, bg)
	}
	cells := []string{
		cellStyled(marker, c.marker, colorAccent, false, false, bg),
		cellStyled(strings.ToUpper(string(st)), c.state, statusColor(st), true, false, bg),
		cellStyled(a.Alarm.Name, c.alarm, rowFg(selected), selected, false, bg),
		signalCell,
		cellStyled(ageShort(a.LastCheckedAt), c.last, colorMuted, false, true, bg),
		trendCell(a.Signals, c.trend, bg),
	}
	sep := sepCell(bg)
	return strings.Join(cells, sep)
}

func rowFg(selected bool) lipgloss.Color {
	if selected {
		return colorBright
	}
	return colorText
}
