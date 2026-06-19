package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/g0ulartleo/mirante/internal/signal"
)

func sparkline(signals []signal.Signal, width int) string {
	if width < 1 {
		return ""
	}
	if len(signals) == 0 {
		return lipgloss.NewStyle().Foreground(colorFaint).Render(strings.Repeat("·", width))
	}
	sl := signals
	if len(sl) > width {
		sl = sl[len(sl)-width:]
	}
	var b strings.Builder
	for _, s := range sl {
		var r string
		switch s.Status {
		case signal.StatusUnhealthy:
			r = "▇"
		case signal.StatusWarning:
			r = "▆"
		case signal.StatusHealthy:
			r = "▂"
		default:
			r = "▁"
		}
		b.WriteString(lipgloss.NewStyle().Foreground(statusColor(s.Status)).Render(r))
	}
	pad := width - len(sl)
	if pad > 0 {
		return lipgloss.NewStyle().Foreground(colorFaint).Render(strings.Repeat("▁", pad)) + b.String()
	}
	return b.String()
}

func histogram(signals []signal.Signal, width, height int, window time.Duration, now time.Time) string {
	if width < 1 || height < 1 {
		return ""
	}
	if window <= 0 {
		window = 24 * time.Hour
	}
	if now.IsZero() {
		now = time.Now()
	}
	start := now.Add(-window)
	cols := make([]signal.Status, width)
	plotWidth := width
	if plotWidth > 1 {
		plotWidth--
	}
	pointSlots := (plotWidth + 1) / 2
	if pointSlots < 1 {
		pointSlots = 1
	}
	for _, s := range signals {
		if s.Timestamp.IsZero() || s.Timestamp.Before(start) {
			continue
		}
		t := s.Timestamp
		if t.After(now) {
			t = now
		}
		age := now.Sub(t)
		slot := pointSlots - 1 - int(age.Seconds()/window.Seconds()*float64(pointSlots))
		if slot < 0 {
			slot = 0
		}
		if slot >= pointSlots {
			slot = pointSlots - 1
		}
		x := slot * 2
		if cols[x] == "" || severityRank(s.Status) > severityRank(cols[x]) {
			cols[x] = s.Status
		}
	}
	rows := make([]string, height)
	barHeight := func(s signal.Status) int {
		switch s {
		case signal.StatusUnhealthy:
			return height
		case signal.StatusWarning:
			return (height + 1) / 2
		case signal.StatusHealthy, signal.StatusUnknown:
			return 1
		default:
			return 0
		}
	}
	for level := 0; level < height; level++ {
		fromBottom := height - level
		var line strings.Builder
		for _, status := range cols {
			h := barHeight(status)
			cell := " "
			if h >= fromBottom {
				cell = "█"
			}
			line.WriteString(lipgloss.NewStyle().Foreground(statusColor(status)).Render(cell))
		}
		rows[level] = line.String()
	}
	return strings.Join(rows, "\n")
}
