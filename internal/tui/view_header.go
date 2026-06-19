package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) headerBand() string {
	switch {
	case m.width < 72:
		return m.summaryPanel(m.width)
	case m.width < 110:
		left := m.contextPanel(28)
		centerW := m.width - lipgloss.Width(left) - 3
		if centerW < 16 {
			centerW = 16
		}
		center := m.summaryPanel(centerW)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, " ", center)
	default:
		left := m.contextPanel(28)
		right := m.keysPanel(30)
		centerW := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 4
		if centerW < 20 {
			centerW = 20
		}
		center := m.summaryPanel(centerW)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, " ", center, " ", right)
	}
}

func (m *Model) contextPanel(w int) string {
	inner := w - 2
	filter := filterLabel(m.filter)
	rows := []string{
		panelTitleStyle.Render("CONTEXT"),
		kv("view", m.listMode.String(), inner),
		kv("filter", truncate(filter, inner-8), inner),
		kv("sort", m.sort.String(), inner),
	}
	return panelBoxStyle.Width(w).Height(4).Render(strings.Join(rows, "\n"))
}

func (m *Model) summaryPanel(w int) string {
	cw := w - 2
	if cw < 1 {
		cw = 1
	}
	hc := m.overallHealth()
	title := lipgloss.NewStyle().Foreground(colorAccent).Render("▌") + brandStyle.Render("mirante")
	total := labelStyle.Render(fmt.Sprintf("%d alarms", hc.total()))
	tgap := cw - lipgloss.Width(title) - lipgloss.Width(total)
	if tgap < 1 {
		tgap = 1
	}
	line1 := title + strings.Repeat(" ", tgap) + total
	rows := []string{
		line1,
		legendItem(colorUnhealthy, hc.unhealthy, "unhealthy"),
		legendItem(colorWarning, hc.warning, "warning"),
		legendItem(colorUnknown, hc.unknown, "unknown"),
		legendItem(colorHealthy, hc.healthy, "healthy"),
		m.healthBar(hc, cw),
	}
	return panelBoxStyle.Width(w).Height(6).Render(strings.Join(rows, "\n"))
}

func legendItem(c lipgloss.Color, n int, label string) string {
	return statusSquare(c) + " " +
		lipgloss.NewStyle().Foreground(colorBright).Bold(true).Render(fmt.Sprintf("%d", n)) + " " +
		labelStyle.Render(label)
}

func (m *Model) healthBar(hc healthCounts, w int) string {
	total := hc.total()
	if total == 0 {
		return lipgloss.NewStyle().Foreground(colorFaint).Render(strings.Repeat("█", w))
	}
	segs := []struct {
		c lipgloss.Color
		n int
	}{
		{colorUnhealthy, hc.unhealthy},
		{colorWarning, hc.warning},
		{colorUnknown, hc.unknown},
		{colorHealthy, hc.healthy},
	}
	var b strings.Builder
	used := 0
	for i, s := range segs {
		var width int
		if i == len(segs)-1 {
			width = w - used
		} else {
			width = s.n * w / total
		}
		if s.n > 0 && width == 0 {
			width = 1
		}
		if used+width > w {
			width = w - used
		}
		if width <= 0 {
			continue
		}
		b.WriteString(lipgloss.NewStyle().Foreground(s.c).Render(strings.Repeat("█", width)))
		used += width
	}
	return b.String()
}

func (m *Model) keysPanel(w int) string {
	inner := w - 2
	rows := []string{
		panelTitleStyle.Render("KEYS"),
		keyRow("j / k", "move cursor", inner),
		keyRow("l / → / ↵", "open alarm", inner),
		keyRow("h / ← / esc", "back · clear", inner),
		keyRow("/ · :", "filter · command", inner),
		keyRow("tab", "toggle view", inner),
		keyRow("?", "help", inner),
	}
	return panelBoxStyle.Width(w).Height(7).Render(strings.Join(rows, "\n"))
}
