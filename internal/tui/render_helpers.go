package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/g0ulartleo/mirante/internal/signal"
)

func kv(key, value string, w int) string {
	k := labelStyle.Render(key)
	v := valueStyle.Render(value)
	gap := w - lipgloss.Width(k) - lipgloss.Width(v)
	if gap < 1 {
		gap = 1
	}
	return k + strings.Repeat(" ", gap) + v
}

func filterLabel(filter string) string {
	if filter == "" {
		return "none"
	}
	return filter
}

func keyRow(key, desc string, w int) string {
	k := lipgloss.NewStyle().Foreground(colorBright).Render(key)
	d := labelStyle.Render(desc)
	gap := w - lipgloss.Width(k) - lipgloss.Width(d)
	if gap < 1 {
		gap = 1
	}
	return k + strings.Repeat(" ", gap) + d
}

func pad(s string, w int, right bool) string {
	s = truncate(s, w)
	style := lipgloss.NewStyle().Width(w)
	if right {
		style = style.Align(lipgloss.Right)
	}
	return style.Render(s)
}

func cellStyled(s string, w int, fg lipgloss.Color, bold, right bool, bg lipgloss.Color) string {
	s = truncate(s, w)
	style := lipgloss.NewStyle().Width(w).Foreground(fg).Bold(bold)
	if right {
		style = style.Align(lipgloss.Right)
	}
	if bg != "" {
		style = style.Background(bg)
	}
	return style.Render(s)
}

func sepCell(bg lipgloss.Color) string {
	st := lipgloss.NewStyle()
	if bg != "" {
		st = st.Background(bg)
	}
	return st.Render(" ")
}

func trendCell(signals []signal.Signal, w int, bg lipgloss.Color) string {
	spark := sparkline(signals, w)
	st := lipgloss.NewStyle().Width(w)
	if bg != "" {
		st = st.Background(bg)
	}
	return st.Render(spark)
}

func sideBySide(left string, rightLines []string, w int) string {
	leftLines := strings.Split(left, "\n")
	n := len(leftLines)
	if len(rightLines) > n {
		n = len(rightLines)
	}
	var out []string
	for i := 0; i < n; i++ {
		l := ""
		if i < len(leftLines) {
			l = leftLines[i]
		}
		r := ""
		if i < len(rightLines) {
			r = rightLines[i]
		}
		gap := w - lipgloss.Width(l) - lipgloss.Width(r)
		if gap < 1 {
			gap = 1
		}
		out = append(out, l+strings.Repeat(" ", gap)+r)
	}
	return strings.Join(out, "\n")
}

func joinWithGap(items []string, gap int) []string {
	if len(items) == 0 {
		return items
	}
	out := make([]string, 0, len(items)*2-1)
	sp := strings.Repeat(" ", gap)
	for i, it := range items {
		if i > 0 {
			out = append(out, sp)
		}
		out = append(out, it)
	}
	return out
}

func renderSignalCell(text string, w int, offset int, bg lipgloss.Color) string {
	if w < 1 {
		return ""
	}
	st := lipgloss.NewStyle().Width(w).Foreground(colorMuted)
	if bg != "" {
		st = st.Background(bg)
	}
	if lipgloss.Width(text) <= w {
		return st.Render(text)
	}
	gap := "   "
	padded := text + gap + text
	start := offset % (len([]rune(text)) + len(gap))
	display := string([]rune(padded)[start : start+w])
	return st.Render(display)
}

func truncate(s string, w int) string {
	if w <= 1 {
		if w == 1 && lipgloss.Width(s) > 0 {
			return "…"
		}
		return ""
	}
	if lipgloss.Width(s) <= w {
		return s
	}
	runes := []rune(s)
	if len(runes) > w-1 {
		runes = runes[:w-1]
	}
	return string(runes) + "…"
}
