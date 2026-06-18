package templates

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/signal"
)

func getGroupStatus(alarmsWithSignals []alarm.AlarmSignals) string {
	if len(alarmsWithSignals) == 0 {
		return "bg-gray-500"
	}
	healthyCount := 0
	warningCount := 0
	for _, a := range alarmsWithSignals {
		if len(a.Signals) == 0 {
			continue
		}
		lastStatus := a.Signals[len(a.Signals)-1].Status
		if lastStatus == "unhealthy" {
			return "bg-red-500"
		}
		if lastStatus == "warning" {
			warningCount++
		}
		if lastStatus == "healthy" {
			healthyCount++
		}
	}
	if warningCount > 0 {
		return "bg-yellow-500"
	}
	if healthyCount == len(alarmsWithSignals) {
		return "bg-green-500"
	}
	return "bg-gray-500"
}

func getAlarmStatusColor(alarmSignals alarm.AlarmSignals) string {
	if len(alarmSignals.Signals) == 0 {
		return "bg-gray-500"
	}
	lastStatus := alarmSignals.Signals[len(alarmSignals.Signals)-1].Status
	if lastStatus == "healthy" {
		return "bg-green-500"
	}
	if lastStatus == "warning" {
		return "bg-yellow-500"
	}
	if lastStatus == "unknown" {
		return "bg-gray-500"
	}
	return "bg-red-500"
}

func getGroupURL(baseURL string, groupKey string) templ.SafeURL {
	if baseURL == "/" {
		return templ.SafeURL(groupKey)
	}
	return templ.SafeURL(fmt.Sprintf("%s/%s", baseURL, groupKey))
}

type treemapGroup struct {
	Name   string
	Alarms []alarm.AlarmSignals
}

type treemapData struct {
	Groups    []treemapGroup
	ThisLevel []alarm.AlarmSignals
	Level     int
	BaseURL   string
}

func buildTreemapData(alarmsWithSignals []alarm.AlarmSignals, level int, baseURL string) treemapData {
	prefix := strings.Trim(strings.TrimLeft(baseURL, "/"), "/")
	var prefixSegments []string
	if prefix == "" {
		prefixSegments = []string{}
	} else {
		prefixSegments = strings.Split(prefix, "/")
	}
	effectiveLevel := len(prefixSegments)

	groups := make(map[string][]alarm.AlarmSignals)
	thisLevel := []alarm.AlarmSignals{}
	for _, a := range alarmsWithSignals {
		pathStr := strings.Join(a.Alarm.Path, "/")

		if prefix == "" {
			if len(a.Alarm.Path) == 0 {
				thisLevel = append(thisLevel, a)
				continue
			}
			key := a.Alarm.Path[0]
			groups[key] = append(groups[key], a)
			continue
		}

		if pathStr == prefix {
			thisLevel = append(thisLevel, a)
			continue
		}

		if strings.HasPrefix(pathStr, prefix+"/") && len(a.Alarm.Path) > effectiveLevel {
			key := a.Alarm.Path[effectiveLevel]
			groups[key] = append(groups[key], a)
		}
	}
	sort.Slice(thisLevel, func(i, j int) bool { return thisLevel[i].Alarm.Name < thisLevel[j].Alarm.Name })
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	outGroups := make([]treemapGroup, 0, len(keys))
	for _, k := range keys {
		gAlarms := groups[k]
		sort.Slice(gAlarms, func(i, j int) bool { return gAlarms[i].Alarm.Name < gAlarms[j].Alarm.Name })
		outGroups = append(outGroups, treemapGroup{Name: k, Alarms: gAlarms})
	}
	return treemapData{Groups: outGroups, ThisLevel: thisLevel, Level: level, BaseURL: baseURL}
}

func getAlarmStatusDotColor(alarmSignals alarm.AlarmSignals) string {
	if len(alarmSignals.Signals) == 0 {
		return "bg-gray-300"
	}
	lastStatus := alarmSignals.Signals[len(alarmSignals.Signals)-1].Status
	if lastStatus == "healthy" {
		return "bg-emerald-300"
	}
	if lastStatus == "warning" {
		return "bg-yellow-300"
	}
	if lastStatus == "unknown" {
		return "bg-gray-300"
	}
	return "bg-red-300"
}

func countUnhealthy(alarmsWithSignals []alarm.AlarmSignals) int {
	c := 0
	for _, a := range alarmsWithSignals {
		if len(a.Signals) == 0 {
			continue
		}
		lastStatus := a.Signals[len(a.Signals)-1].Status
		if lastStatus == "unhealthy" {
			c++
		}
	}
	return c
}

func sumUnhealthy24h(alarmsWithSignals []alarm.AlarmSignals) int {
	total := 0
	for _, a := range alarmsWithSignals {
		total += a.UnhealthyCount24h
	}
	return total
}

func lastCheckedGroup(alarmsWithSignals []alarm.AlarmSignals) time.Time {
	var latest time.Time
	for _, a := range alarmsWithSignals {
		if a.LastCheckedAt.After(latest) {
			latest = a.LastCheckedAt
		}
	}
	return latest
}

func humanizeDuration(d time.Duration) string {
	seconds := int(d.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	if hours < 24 {
		return fmt.Sprintf("%dh", hours)
	}
	days := hours / 24
	return fmt.Sprintf("%dd", days)
}

func formatSignalDetails(sig signal.Signal) string {
	if len(sig.Details) == 0 {
		return ""
	}
	var b strings.Builder
	for i, detail := range sig.Details {
		if i > 0 {
			b.WriteString("\n")
		}
		if detail.Title != "" {
			b.WriteString(detail.Title)
			b.WriteString(": ")
		}
		b.WriteString(formatSignalDetailValue(detail))
	}
	return b.String()
}

func formatSignalDetailValue(detail signal.Detail) string {
	switch detail.Type {
	case signal.DetailTypeText:
		return detail.Text
	case signal.DetailTypeObject:
		data, err := json.Marshal(detail.Object)
		if err != nil {
			return fmt.Sprintf("%v", detail.Object)
		}
		return string(data)
	case signal.DetailTypeTable:
		if detail.Table == nil {
			return ""
		}
		lines := []string{}
		if len(detail.Table.Columns) > 0 {
			lines = append(lines, strings.Join(detail.Table.Columns, " | "))
		}
		for _, row := range detail.Table.Rows {
			lines = append(lines, strings.Join(row, " | "))
		}
		return strings.Join(lines, "\n")
	case signal.DetailTypeList:
		return strings.Join(detail.List, ", ")
	default:
		return ""
	}
}
