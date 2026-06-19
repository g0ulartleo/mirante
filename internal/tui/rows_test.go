package tui

import (
	"testing"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/signal"
)

func TestBuildRowsSortsAlarmsWithinTopLevelGroup(t *testing.T) {
	rows, alarmRows := buildRows([]alarm.AlarmSignals{
		alarmSignals("2", "Zulu", []string{"api", "z"}, signal.StatusHealthy),
		alarmSignals("1", "Alpha", []string{"api", "a"}, signal.StatusHealthy),
	}, "", sortName)

	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
	if rows[0].kind != rowGroupHeader || rows[0].groupName != "api" {
		t.Fatalf("expected api group header, got %#v", rows[0])
	}
	if len(alarmRows) != 2 || alarmRows[0] != 1 || alarmRows[1] != 2 {
		t.Fatalf("unexpected alarm rows: %#v", alarmRows)
	}
	if rows[1].alarm.Alarm.Name != "Alpha" || rows[2].alarm.Alarm.Name != "Zulu" {
		t.Fatalf("alarms not sorted by name: %q, %q", rows[1].alarm.Alarm.Name, rows[2].alarm.Alarm.Name)
	}
}

func TestBuildRowsSortsGroupsByAttentionThenName(t *testing.T) {
	rows, _ := buildRows([]alarm.AlarmSignals{
		alarmSignals("1", "Healthy", []string{"api"}, signal.StatusHealthy),
		alarmSignals("2", "Broken", []string{"worker"}, signal.StatusUnhealthy),
		alarmSignals("3", "Warn", []string{"db"}, signal.StatusWarning),
	}, "", sortSeverity)

	groups := []string{rows[0].groupName, rows[2].groupName, rows[4].groupName}
	want := []string{"db", "worker", "api"}
	for i := range want {
		if groups[i] != want[i] {
			t.Fatalf("group order = %#v, want %#v", groups, want)
		}
	}
}

func TestMatchesFilterAttentionKeyword(t *testing.T) {
	if !matchesFilter(alarmSignals("1", "Warn", []string{"api"}, signal.StatusWarning), "attention") {
		t.Fatal("warning alarm should match attention")
	}
	if matchesFilter(alarmSignals("2", "Healthy", []string{"api"}, signal.StatusHealthy), "needs-attention") {
		t.Fatal("healthy alarm should not match needs-attention")
	}
}

func alarmSignals(id, name string, path []string, status signal.Status) alarm.AlarmSignals {
	return alarm.AlarmSignals{
		Alarm: alarm.Alarm{
			ID:   id,
			Name: name,
			Path: path,
		},
		Signals: []signal.Signal{{Status: status}},
	}
}
