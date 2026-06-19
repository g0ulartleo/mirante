package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/g0ulartleo/mirante/internal/signal"
)

func (m *Model) renderHistoryPanel(w int) string {
	inner := w - panelBoxStyle.GetHorizontalFrameSize()
	if inner < 10 {
		inner = 10
	}
	boxW := w - panelBoxStyle.GetHorizontalBorderSize()
	if boxW < inner {
		boxW = inner
	}
	now := m.detailHistoryAt
	if now.IsZero() {
		now = time.Now()
	}
	window := historyWindow(m.detailSignals, now)
	titleText := "SIGNAL HISTORY · " + durationLabel(window)
	if stale := historyStaleLabel(now); stale != "" {
		titleText += " ending " + stale
	}
	title := panelTitleStyle.Render(titleText)
	legend := statusSquare(colorHealthy) + labelStyle.Render(" healthy  ") + statusSquare(colorUnhealthy) + labelStyle.Render(" unhealthy")
	tgap := inner - lipgloss.Width(title) - lipgloss.Width(legend)
	if tgap < 1 {
		tgap = 1
	}
	head := title + strings.Repeat(" ", tgap) + legend
	var chart string
	if m.detailLoading {
		chart = labelStyle.Render("loading…")
	} else if len(m.detailSignals) == 0 {
		chart = labelStyle.Render("no signal history")
	} else {
		chart = histogram(m.detailSignals, inner, 4, window, now)
		axis := m.timeAxis(inner, window, now)
		chart = chart + "\n" + axis
	}
	body := head + "\n\n" + chart
	return panelBoxStyle.Width(boxW).Render(body)
}

func historyWindow(signals []signal.Signal, now time.Time) time.Duration {
	if now.IsZero() {
		now = time.Now()
	}
	oldest := now
	found := false
	for _, s := range signals {
		if s.Timestamp.IsZero() {
			continue
		}
		if s.Timestamp.Before(oldest) {
			oldest = s.Timestamp
		}
		found = true
	}
	if !found {
		return time.Minute
	}
	span := now.Sub(oldest)
	if span < time.Minute {
		return time.Minute
	}
	return span
}

func historyStaleLabel(anchor time.Time) string {
	if anchor.IsZero() {
		return ""
	}
	age := time.Since(anchor)
	if age < 5*time.Minute {
		return ""
	}
	return ageShort(anchor) + " ago"
}

func durationLabel(d time.Duration) string {
	if d >= 24*time.Hour {
		days := int(d / (24 * time.Hour))
		hours := int((d % (24 * time.Hour)) / time.Hour)
		if hours == 0 {
			return fmt.Sprintf("%dd", days)
		}
		return fmt.Sprintf("%dd%dh", days, hours)
	}
	if d >= time.Hour {
		hours := int(d / time.Hour)
		minutes := int((d % time.Hour) / time.Minute)
		if minutes == 0 {
			return fmt.Sprintf("%dh", hours)
		}
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", int(d/time.Minute))
}

func (m *Model) timeAxis(w int, window time.Duration, anchor time.Time) string {
	if anchor.IsZero() {
		anchor = time.Now()
	}
	labels := []string{
		historyAxisLabel(anchor.Add(-window), window, anchor),
		historyAxisLabel(anchor.Add(-window*3/4), window, anchor),
		historyAxisLabel(anchor.Add(-window/2), window, anchor),
		historyAxisLabel(anchor.Add(-window/4), window, anchor),
		historyLastLabel(anchor),
	}
	seg := w / (len(labels) - 1)
	var b strings.Builder
	for i, l := range labels {
		if i == 0 {
			b.WriteString(l)
		} else {
			pad := seg - len(labels[i-1])
			if i == len(labels)-1 {
				rem := w - lipgloss.Width(b.String()) - len(l)
				if rem < 1 {
					rem = 1
				}
				b.WriteString(strings.Repeat(" ", rem) + l)
				continue
			}
			if pad < 1 {
				pad = 1
			}
			b.WriteString(strings.Repeat(" ", pad) + l)
		}
	}
	return labelStyle.Render(b.String())
}

func historyAxisLabel(t time.Time, window time.Duration, anchor time.Time) string {
	if window >= 24*time.Hour {
		return t.Format("Jan 2")
	}
	if sameDay(t, time.Now()) && sameDay(anchor, time.Now()) {
		return t.Format("15:04")
	}
	return t.Format("Jan2 15:04")
}

func historyLastLabel(anchor time.Time) string {
	if anchor.IsZero() || time.Since(anchor) < 5*time.Minute {
		return "last"
	}
	return "last " + ageShort(anchor)
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
