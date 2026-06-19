package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/signal"
)

func (m *Model) detailScreen() string {
	as, ok := m.selectedAlarm()
	crumb := m.detailBreadcrumb(as, ok)
	status := m.statusBar()
	return lipgloss.JoinVertical(lipgloss.Left, crumb, m.detail.View(), status)
}

func (m *Model) detailBreadcrumb(as alarm.AlarmSignals, ok bool) string {
	parts := []string{"mirante", m.listMode.String()}
	if ok {
		parts = append(parts, as.Alarm.Path...)
		parts = append(parts, as.Alarm.Name)
	}
	left := lipgloss.NewStyle().Foreground(colorAccent).Render("← ")
	rendered := make([]string, len(parts))
	for i, p := range parts {
		if i == len(parts)-1 {
			rendered[i] = crumbActiveStyle.Render(p)
		} else {
			rendered[i] = breadcrumbStyle.Render(p)
		}
	}
	left += strings.Join(rendered, breadcrumbStyle.Render(" › "))
	right := breadcrumbStyle.Render("h back")
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	gap = max(gap, 1)
	return left + strings.Repeat(" ", gap) + right
}

func (m *Model) renderDetail(as alarm.AlarmSignals) string {
	w := m.width - 4
	w = max(w, 20)
	compact := m.detail.Height < 32
	var b strings.Builder
	st := lastStatus(as)
	titleLeft := detailTitleStyle.Render(as.Alarm.Name) + "  " + statusBadgeBox(st)
	right := labelStyle.Render("path ") + valueStyle.Render(strings.Join(as.Alarm.Path, "/"))
	titleBlock := sideBySide(titleLeft, strings.Split(right, "\n"), w)
	b.WriteString(titleBlock)
	b.WriteString("\n")
	b.WriteString(valueStyle.Render(as.Alarm.ID) + " (y to copy) \n\n")
	if as.Alarm.Description != "" {
		b.WriteString(labelStyle.Width(w).Render(as.Alarm.Description))
		b.WriteString("\n")
	}
	if as.Alarm.HowToFix != "" {
		b.WriteString(labelStyle.Width(w).Render(as.Alarm.HowToFix))
		b.WriteString("\n")
	}
	if !compact {
		b.WriteString("\n")
		b.WriteString(m.renderCards(as, w))
		b.WriteString("\n")
		b.WriteString(m.renderHistoryPanel(w))
		b.WriteString("\n\n")
	} else {
		b.WriteString("\n")
	}
	b.WriteString(m.renderSignalsTable(w))
	b.WriteString("\n")
	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func statusBadgeBox(s signal.Status) string {
	text := strings.ToUpper(string(s))
	if text == "" {
		text = "UNKNOWN"
	}
	return lipgloss.NewStyle().Foreground(colorBg).Background(statusColor(s)).Bold(true).Padding(0, 1).Render(text)
}

func (m *Model) renderCards(as alarm.AlarmSignals, w int) string {
	gap := 1
	border := cardStyle.GetHorizontalBorderSize()
	total := w - 2*gap - 3*border
	if total < 3 {
		total = 3
	}
	cardW := total / 3
	rem := total % 3
	cw1 := cardW
	if rem > 0 {
		cw1++
		rem--
	}
	cw2 := cardW
	if rem > 0 {
		cw2++
		rem--
	}
	cw3 := cardW
	interval := as.Alarm.Interval
	if interval == "" {
		interval = as.Alarm.Cron
	}
	if interval == "" {
		interval = "30s"
	}
	cards := []string{
		card("UNHEALTHY · 24H", lipgloss.NewStyle().Foreground(colorUnhealthy).Bold(true).Render(fmt.Sprintf("%d", as.UnhealthyCount24h))+labelStyle.Render(" signals"), cw1),
		card("LAST CHECKED", valueStyle.Bold(true).Render(ageShort(as.LastCheckedAt))+labelStyle.Render(" ago"), cw2),
		card("INTERVAL", valueStyle.Render(interval), cw3),
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, joinWithGap(cards, gap)...)
}

func card(title, value string, w int) string {
	body := panelTitleStyle.Render(title) + "\n\n" + value
	return cardStyle.Width(w).Render(body)
}

func (m *Model) renderSignalsTable(w int) string {
	inner := w - panelBoxStyle.GetHorizontalFrameSize()
	if inner < 20 {
		inner = 20
	}
	boxW := w - panelBoxStyle.GetHorizontalBorderSize()
	if boxW < inner {
		boxW = inner
	}
	numW := 3
	statusW := 10
	timeW := 14
	sep := 1
	fixed := numW + statusW + timeW + 3*sep
	resW := inner - fixed
	if resW < 20 {
		resW = 20
	}
	count := len(m.detailSignals)
	signalCursor := m.detailSignalCursor
	if count > 0 && signalCursor >= count {
		signalCursor = count - 1
	}
	page := 0
	pageCount := 0
	start := 0
	end := count
	if count > 0 {
		pageCount = (count + detailSignalsPageSize - 1) / detailSignalsPageSize
		page = signalCursor / detailSignalsPageSize
		start = page * detailSignalsPageSize
		end = start + detailSignalsPageSize
		if end > count {
			end = count
		}
	}
	titleText := fmt.Sprintf("LAST SIGNALS (%d)", count)
	if pageCount > 1 {
		titleText = fmt.Sprintf("%s · page %d/%d", titleText, page+1, pageCount)
	}
	title := panelTitleStyle.Render(titleText)
	hint := labelStyle.Render("· j/k · [ ] page · c copy")
	tgap := inner - lipgloss.Width(title) - lipgloss.Width(hint)
	if tgap < 1 {
		tgap = 1
	}
	head := title + strings.Repeat(" ", tgap) + hint
	spc := strings.Repeat(" ", sep)
	header := colHeaderStyle.Render(
		pad("#", numW, true) + spc +
			pad("STATUS", statusW, false) + spc +
			pad("TIME", timeW, false) + spc +
			pad("RESPONSE", resW, false),
	)
	var rows []string
	rows = append(rows, head, "", header)
	if m.detailLoading {
		rows = append(rows, labelStyle.Render("loading…"))
	} else if count == 0 {
		rows = append(rows, labelStyle.Render("no signals"))
	} else {
		for i := start; i < end; i++ {
			s := m.detailSignals[i]
			isCursor := i == signalCursor
			var bg lipgloss.Color
			if isCursor {
				bg = colorSelBg
			}
			num := fmt.Sprintf("%d", i+1)
			if isCursor {
				num = "▶" + num
			} else {
				num = " " + num
			}
			numC := cellStyled(num, numW, colorAccent, false, false, bg)
			statusCell := cellStyled(strings.ToUpper(string(s.Status)), statusW, statusColor(s.Status), true, false, bg)
			timeCell := cellStyled(ageShort(s.Timestamp)+" ago", timeW, colorMuted, false, false, bg)
			resCell := cellStyled(s.Message, resW, colorText, false, false, bg)
			rowSep := spc
			if bg != "" {
				rowSep = sepCell(bg)
			}
			rows = append(rows, numC+rowSep+statusCell+rowSep+timeCell+rowSep+resCell)
			if isCursor {
				expandStyle := lipgloss.NewStyle().Foreground(colorMuted).Width(inner)
				detail := strings.TrimRight(signalDetailText(s), "\n")
				rows = append(rows, strings.Split(expandStyle.Render(detail), "\n")...)
			}
		}
	}
	return panelBoxStyle.Width(boxW).Render(strings.Join(rows, "\n"))
}
