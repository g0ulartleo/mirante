package main

import (
	"testing"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/worker/tasks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSchedulerAlarmRepo struct {
	alarms []*alarm.Alarm
}

func (r *fakeSchedulerAlarmRepo) Init() error { return nil }

func (r *fakeSchedulerAlarmRepo) GetAlarms() ([]*alarm.Alarm, error) {
	return r.alarms, nil
}

func (r *fakeSchedulerAlarmRepo) GetAlarm(alarmID string) (*alarm.Alarm, error) { return nil, nil }

func (r *fakeSchedulerAlarmRepo) SetAlarm(alarm *alarm.Alarm) error { return nil }

func (r *fakeSchedulerAlarmRepo) DeleteAlarm(alarmID string) error { return nil }

func (r *fakeSchedulerAlarmRepo) DeleteStaleAlarmsByRuntime(runtime string, keepIDs map[string]bool) error {
	return nil
}

func (r *fakeSchedulerAlarmRepo) Close() error { return nil }

func TestAlarmConfigProviderCreatesTasksFromSyncedAlarms(t *testing.T) {
	repo := &fakeSchedulerAlarmRepo{alarms: []*alarm.Alarm{
		{ID: "interval-alarm", Interval: "1m", Runtime: "go"},
		{ID: "cron-alarm", Cron: "*/5 * * * *", Runtime: "node"},
	}}
	provider := &AlarmConfigProvider{alarmService: alarm.NewAlarmService(repo)}

	configs, err := provider.GetConfigs()
	require.NoError(t, err)
	require.Len(t, configs, 3)

	assert.Equal(t, "@every 1m", configs[0].Cronspec)
	assert.Equal(t, tasks.TypeAlarmCheck, configs[0].Task.Type())
	assert.Equal(t, "*/5 * * * *", configs[1].Cronspec)
	assert.Equal(t, tasks.TypeAlarmCheck, configs[1].Task.Type())
	assert.Equal(t, "@every 24h", configs[2].Cronspec)
	assert.Equal(t, tasks.TypeBackofficeCleanSignals, configs[2].Task.Type())
}
