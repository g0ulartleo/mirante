package tasks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/g0ulartleo/mirante/internal/alarm"
	runtimeclient "github.com/g0ulartleo/mirante/internal/alarm/runtime/client"
	"github.com/g0ulartleo/mirante/internal/signal"
	alarmsv1 "github.com/g0ulartleo/mirante/proto/alarms/v1"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRunner struct {
	signal      signal.Signal
	err         error
	runtimeName string
	alarmID     string
}

func (f *fakeRunner) RunAlarm(ctx context.Context, runtimeName string, alarmID string) (signal.Signal, error) {
	f.runtimeName = runtimeName
	f.alarmID = alarmID
	return f.signal, f.err
}

type fakeEnqueuer struct {
	tasks []*asynq.Task
}

func (f *fakeEnqueuer) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	f.tasks = append(f.tasks, task)
	return &asynq.TaskInfo{}, nil
}

type fakeAlarmRepo struct {
	alarm *alarm.Alarm
	err   error
}

func (f *fakeAlarmRepo) Init() error { return nil }

func (f *fakeAlarmRepo) GetAlarms() ([]*alarm.Alarm, error) { return nil, nil }

func (f *fakeAlarmRepo) GetAlarm(alarmID string) (*alarm.Alarm, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.alarm, nil
}

func (f *fakeAlarmRepo) SetAlarm(alarm *alarm.Alarm) error { return nil }

func (f *fakeAlarmRepo) DeleteAlarm(alarmID string) error { return nil }

func (f *fakeAlarmRepo) DeleteStaleAlarmsByRuntime(runtime string, keepIDs map[string]bool) error {
	return nil
}

func (f *fakeAlarmRepo) Close() error { return nil }

type fakeSignalRepo struct {
	saved []signal.Signal
}

func (f *fakeSignalRepo) Init() error { return nil }

func (f *fakeSignalRepo) Close() error { return nil }

func (f *fakeSignalRepo) Save(sig signal.Signal) error {
	f.saved = append(f.saved, sig)
	return nil
}

func (f *fakeSignalRepo) GetAlarmLatestSignals(alarmID string, limit int) ([]signal.Signal, error) {
	if len(f.saved) == 0 {
		return nil, nil
	}
	if limit <= 0 || limit >= len(f.saved) {
		return append([]signal.Signal(nil), f.saved...), nil
	}
	return append([]signal.Signal(nil), f.saved[len(f.saved)-limit:]...), nil
}

func (f *fakeSignalRepo) GetAlarmSignalsSince(alarmID string, since time.Time) ([]signal.Signal, error) {
	matched := make([]signal.Signal, 0, len(f.saved))
	for _, s := range f.saved {
		if s.Timestamp.After(since) || s.Timestamp.Equal(since) {
			matched = append(matched, s)
		}
	}
	return matched, nil
}

func (f *fakeSignalRepo) GetAlarmHealth(alarmID string) (signal.Status, error) {
	if len(f.saved) == 0 {
		return signal.StatusUnknown, nil
	}
	return f.saved[len(f.saved)-1].Status, nil
}

func (f *fakeSignalRepo) CountUnhealthySince(alarmID string, since time.Time) (int, error) {
	return 0, nil
}

func (f *fakeSignalRepo) CleanOldSignals() error { return nil }

func TestCheckAlarm_UnsupportedReturnsSkipRetry(t *testing.T) {
	alarmService := alarm.NewAlarmService(&fakeAlarmRepo{alarm: &alarm.Alarm{
		ID: "alarm-1",
	}})
	signalRepo := &fakeSignalRepo{}
	signalService := signal.NewService(signalRepo)

	runtimeErr := &runtimeclient.RuntimeError{
		Code: alarmsv1.RuntimeErrorCode_RUNTIME_ERROR_CODE_UNSUPPORTED,
		Signal: signal.Signal{
			AlarmID:   "alarm-1",
			Status:    signal.StatusUnknown,
			Timestamp: time.Now(),
			Message:   "not found",
		},
	}

	err := checkAlarm(
		context.Background(),
		AlarmCheckPayload{AlarmID: "alarm-1"},
		&fakeRunner{err: runtimeErr},
		signalService,
		alarmService,
		nil,
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, asynq.SkipRetry)
	require.Len(t, signalRepo.saved, 1)
	assert.Equal(t, signal.StatusUnknown, signalRepo.saved[0].Status)
	assert.Equal(t, "not found", signalRepo.saved[0].Message)
}

func TestCheckAlarm_CheckFailedWritesUnknownAndReturnsNil(t *testing.T) {
	alarmService := alarm.NewAlarmService(&fakeAlarmRepo{alarm: &alarm.Alarm{
		ID: "alarm-1",
	}})
	signalRepo := &fakeSignalRepo{}
	signalService := signal.NewService(signalRepo)

	runtimeErr := &runtimeclient.RuntimeError{
		Code: alarmsv1.RuntimeErrorCode_RUNTIME_ERROR_CODE_RUNALARM_FAILED,
		Signal: signal.Signal{
			AlarmID:   "alarm-1",
			Status:    signal.StatusUnknown,
			Timestamp: time.Now(),
			Message:   "temporary check failure",
		},
	}

	err := checkAlarm(
		context.Background(),
		AlarmCheckPayload{AlarmID: "alarm-1"},
		&fakeRunner{err: runtimeErr},
		signalService,
		alarmService,
		nil,
	)

	require.NoError(t, err)
	require.Len(t, signalRepo.saved, 1)
	assert.Equal(t, signal.StatusUnknown, signalRepo.saved[0].Status)
	assert.Equal(t, "temporary check failure", signalRepo.saved[0].Message)
}

func TestCheckAlarm_TransportErrorReturnsErrorWithoutSignalWrite(t *testing.T) {
	alarmService := alarm.NewAlarmService(&fakeAlarmRepo{alarm: &alarm.Alarm{
		ID: "alarm-1",
	}})
	signalRepo := &fakeSignalRepo{}
	signalService := signal.NewService(signalRepo)

	err := checkAlarm(
		context.Background(),
		AlarmCheckPayload{AlarmID: "alarm-1"},
		&fakeRunner{err: errors.New("rpc unavailable")},
		signalService,
		alarmService,
		nil,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check alarm via runtime")
	assert.Empty(t, signalRepo.saved)
}

func TestCheckAlarm_WritesRuntimeStatusSignals(t *testing.T) {
	tests := []signal.Status{
		signal.StatusHealthy,
		signal.StatusWarning,
		signal.StatusUnhealthy,
		signal.StatusUnknown,
	}

	for _, status := range tests {
		t.Run(string(status), func(t *testing.T) {
			alarmService := alarm.NewAlarmService(&fakeAlarmRepo{alarm: &alarm.Alarm{
				ID:      "alarm-1",
				Runtime: "go",
			}})
			signalRepo := &fakeSignalRepo{}
			signalService := signal.NewService(signalRepo)
			runner := &fakeRunner{signal: signal.Signal{
				AlarmID:   "alarm-1",
				Status:    status,
				Timestamp: time.Now(),
				Message:   "runtime response",
			}}
			enqueuer := &fakeEnqueuer{}

			err := checkAlarm(
				context.Background(),
				AlarmCheckPayload{AlarmID: "alarm-1"},
				runner,
				signalService,
				alarmService,
				enqueuer,
			)

			require.NoError(t, err)
			assert.Equal(t, "go", runner.runtimeName)
			assert.Equal(t, "alarm-1", runner.alarmID)
			require.Len(t, signalRepo.saved, 1)
			assert.Equal(t, status, signalRepo.saved[0].Status)
			assert.Equal(t, "runtime response", signalRepo.saved[0].Message)
			require.Len(t, enqueuer.tasks, 1)
			assert.Equal(t, TypeDashboardNotify, enqueuer.tasks[0].Type())
		})
	}
}
