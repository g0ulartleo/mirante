package alarms

import (
	"context"
	"testing"

	alarmsv1 "github.com/g0ulartleo/mirante/packages/go/alarms/v1"
)

func TestRuntimeServerListsAndRunsAlarms(t *testing.T) {
	server, err := NewRuntimeServer([]Alarm{{
		ID:          "ping-api",
		Name:        "Ping API",
		Description: "Pings an API",
		Interval:    "1m",
		Run: func(ctx context.Context) (*Signal, error) {
			return Healthy("OK"), nil
		},
	}})
	if err != nil {
		t.Fatalf("NewRuntimeServer() error = %v", err)
	}

	listed, err := server.ListAlarms(context.Background(), &alarmsv1.ListAlarmsRequest{})
	if err != nil {
		t.Fatalf("ListAlarms() error = %v", err)
	}
	if got := len(listed.GetAlarms()); got != 1 {
		t.Fatalf("len(ListAlarms()) = %d, want 1", got)
	}
	if got := listed.GetAlarms()[0].GetId(); got != "ping-api" {
		t.Fatalf("alarm id = %q, want ping-api", got)
	}

	run, err := server.RunAlarm(context.Background(), &alarmsv1.RunAlarmRequest{AlarmId: "ping-api"})
	if err != nil {
		t.Fatalf("RunAlarm() error = %v", err)
	}
	if got := run.GetStatus(); got != alarmsv1.SignalStatus_SIGNAL_STATUS_HEALTHY {
		t.Fatalf("status = %s, want healthy", got)
	}
}

func TestRuntimeServerRejectsInvalidAlarm(t *testing.T) {
	_, err := NewRuntimeServer([]Alarm{{ID: "missing-fields"}})
	if err == nil {
		t.Fatal("NewRuntimeServer() error = nil, want error")
	}
}
