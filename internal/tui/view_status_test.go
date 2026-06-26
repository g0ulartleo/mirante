package tui

import (
	"strings"
	"testing"

	"github.com/g0ulartleo/mirante/internal/signal"
)

func TestStatusBadgeModeAllUsesAccent(t *testing.T) {
	m := &Model{listMode: allList}
	mode, color := m.statusBadgeMode()
	if mode != "ALL" || color != colorAccent {
		t.Fatalf("badge = %q %q, want ALL %q", mode, color, colorAccent)
	}
}

func TestStatusBarShowsFilterWhenFilterApplied(t *testing.T) {
	m := &Model{listMode: triageList, filter: "api", width: 80, connected: true}
	status := m.statusBar()
	if !strings.Contains(status, "FILTER") {
		t.Fatalf("status bar = %q, want FILTER badge", status)
	}
}

func TestStatusBadgeModeTriageHealthColors(t *testing.T) {
	tests := []struct {
		name  string
		data  []signal.Status
		color string
	}{
		{name: "unhealthy", data: []signal.Status{signal.StatusHealthy, signal.StatusUnhealthy}, color: string(colorUnhealthy)},
		{name: "unknown", data: []signal.Status{signal.StatusHealthy, signal.StatusUnknown}, color: string(colorUnknown)},
		{name: "warning is green", data: []signal.Status{signal.StatusHealthy, signal.StatusWarning}, color: string(colorHealthy)},
		{name: "healthy", data: []signal.Status{signal.StatusHealthy}, color: string(colorHealthy)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Model{listMode: triageList}
			for i, st := range tt.data {
				m.all = append(m.all, alarmSignals(string(rune('a'+i)), string(rune('a'+i)), nil, st))
			}
			mode, color := m.statusBadgeMode()
			if mode != "TRIAGE" || string(color) != tt.color {
				t.Fatalf("badge = %q %q, want TRIAGE %q", mode, color, tt.color)
			}
		})
	}
}
