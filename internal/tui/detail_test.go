package tui

import (
	"testing"
	"time"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/signal"
)

func TestMergeDetailSignalsKeepsNewestFirstAndDedupes(t *testing.T) {
	base := time.Date(2026, 6, 26, 18, 0, 0, 0, time.UTC)
	old := detailTestSignal("a", signal.StatusUnknown, base.Add(-4*time.Minute), "old")
	latest := detailTestSignal("a", signal.StatusHealthy, base, "latest")
	middle := detailTestSignal("a", signal.StatusUnhealthy, base.Add(-2*time.Minute), "middle")

	got := mergeDetailSignals([]signal.Signal{middle, old}, []signal.Signal{latest, middle}, 0)

	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	wantMessages := []string{"latest", "middle", "old"}
	for i, want := range wantMessages {
		if got[i].Message != want {
			t.Fatalf("messages = %#v, want %#v", []string{got[0].Message, got[1].Message, got[2].Message}, wantMessages)
		}
	}

	limited := mergeDetailSignals(got, nil, 2)
	if len(limited) != 2 || limited[0].Message != "latest" || limited[1].Message != "middle" {
		t.Fatalf("limited = %#v, want latest/middle", limited)
	}
}

func TestApplyDataRefreshesSelectedDetailSignals(t *testing.T) {
	base := time.Date(2026, 6, 26, 18, 0, 0, 0, time.UTC)
	old := detailTestSignal("alarm-1", signal.StatusUnknown, base.Add(-2*time.Minute), "old")
	fresh := detailTestSignal("alarm-1", signal.StatusHealthy, base, "fresh")
	m := &Model{
		mode:          detailView,
		selectedID:    "alarm-1",
		detailSignals: []signal.Signal{old},
		width:         80,
		height:        40,
	}
	m.resizeDetail()

	m.applyData([]alarm.AlarmSignals{{
		Alarm:         alarm.Alarm{ID: "alarm-1", Name: "Alarm"},
		Signals:       []signal.Signal{fresh},
		LastCheckedAt: fresh.Timestamp,
	}})

	if len(m.detailSignals) != 2 {
		t.Fatalf("detailSignals len = %d, want 2", len(m.detailSignals))
	}
	if m.detailSignals[0].Message != "fresh" {
		t.Fatalf("latest detail signal = %q, want fresh", m.detailSignals[0].Message)
	}
	if !m.detailHistoryAt.Equal(fresh.Timestamp) {
		t.Fatalf("detailHistoryAt = %s, want %s", m.detailHistoryAt, fresh.Timestamp)
	}
}

func detailTestSignal(alarmID string, status signal.Status, ts time.Time, message string) signal.Signal {
	return signal.Signal{AlarmID: alarmID, Status: status, Timestamp: ts, Message: message}
}
