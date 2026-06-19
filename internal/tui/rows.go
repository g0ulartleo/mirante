package tui

import (
	"sort"
	"strings"

	"github.com/g0ulartleo/mirante/internal/alarm"
)

type rowKind int

const (
	rowGroupHeader rowKind = iota
	rowAlarm
)

type listRow struct {
	kind        rowKind
	groupName   string
	groupHealth healthCounts
	alarm       alarm.AlarmSignals
}

type sortMode int

const (
	sortSeverity sortMode = iota
	sortName
)

func (s sortMode) String() string {
	switch s {
	case sortName:
		return "name ↓"
	default:
		return "severity ↓"
	}
}

func groupKey(a alarm.AlarmSignals) string {
	if len(a.Alarm.Path) == 0 {
		return "ungrouped"
	}
	return a.Alarm.Path[0]
}

func matchesFilter(a alarm.AlarmSignals, filter string) bool {
	f := strings.TrimSpace(strings.ToLower(filter))
	if f == "" {
		return true
	}
	if f == "needs-attention" || f == "attention" {
		return severityRank(lastStatus(a)) >= 2
	}
	hay := strings.ToLower(a.Alarm.Name + " " + groupKey(a) + " " + string(lastStatus(a)))
	if len(a.Signals) > 0 {
		hay += " " + strings.ToLower(a.Signals[len(a.Signals)-1].Message)
	}
	return strings.Contains(hay, f)
}

func buildRows(all []alarm.AlarmSignals, filter string, mode sortMode) ([]listRow, []int) {
	groups := make(map[string][]alarm.AlarmSignals)
	var order []string
	for _, a := range all {
		if !matchesFilter(a, filter) {
			continue
		}
		k := groupKey(a)
		if _, ok := groups[k]; !ok {
			order = append(order, k)
		}
		groups[k] = append(groups[k], a)
	}

	for k := range groups {
		g := groups[k]
		sort.Slice(g, func(i, j int) bool {
			return lessAlarm(g[i], g[j], mode)
		})
		groups[k] = g
	}

	sort.Slice(order, func(i, j int) bool {
		if mode == sortSeverity {
			ai := countHealth(groups[order[i]]).needAttention()
			aj := countHealth(groups[order[j]]).needAttention()
			if ai != aj {
				return ai > aj
			}
		}
		return order[i] < order[j]
	})

	var rows []listRow
	var alarmRows []int
	for _, k := range order {
		g := groups[k]
		rows = append(rows, listRow{
			kind:        rowGroupHeader,
			groupName:   k,
			groupHealth: countHealth(g),
		})
		for _, a := range g {
			alarmRows = append(alarmRows, len(rows))
			rows = append(rows, listRow{kind: rowAlarm, groupName: k, alarm: a})
		}
	}
	return rows, alarmRows
}

func lessAlarm(a, b alarm.AlarmSignals, mode sortMode) bool {
	if mode == sortSeverity {
		ra, rb := severityRank(lastStatus(a)), severityRank(lastStatus(b))
		if ra != rb {
			return ra > rb
		}
	}
	return a.Alarm.Name < b.Alarm.Name
}
