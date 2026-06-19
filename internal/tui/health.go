package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/signal"
)

type healthCounts struct {
	unhealthy int
	warning   int
	unknown   int
	healthy   int
}

func statusColor(s signal.Status) lipgloss.Color {
	switch s {
	case signal.StatusHealthy:
		return colorHealthy
	case signal.StatusWarning:
		return colorWarning
	case signal.StatusUnhealthy:
		return colorUnhealthy
	default:
		return colorUnknown
	}
}

func lastStatus(as alarm.AlarmSignals) signal.Status {
	if len(as.Signals) == 0 {
		return ""
	}
	return as.Signals[len(as.Signals)-1].Status
}

func statusSquare(c lipgloss.Color) string {
	return lipgloss.NewStyle().Foreground(c).Render("■")
}

func (h healthCounts) total() int {
	return h.unhealthy + h.warning + h.unknown + h.healthy
}

func (h healthCounts) needAttention() int {
	return h.unhealthy + h.warning
}

func countHealth(group []alarm.AlarmSignals) healthCounts {
	var hc healthCounts
	for _, a := range group {
		switch lastStatus(a) {
		case signal.StatusUnhealthy:
			hc.unhealthy++
		case signal.StatusWarning:
			hc.warning++
		case signal.StatusHealthy:
			hc.healthy++
		default:
			hc.unknown++
		}
	}
	return hc
}

func severityRank(s signal.Status) int {
	switch s {
	case signal.StatusUnhealthy:
		return 3
	case signal.StatusWarning:
		return 2
	case signal.StatusUnknown, "":
		return 1
	default:
		return 0
	}
}
