package tasks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/g0ulartleo/mirante/internal/alarm"
	runtimeclient "github.com/g0ulartleo/mirante/internal/sentinel/runtime/client"
	"github.com/g0ulartleo/mirante/internal/signal"
	runtimev1 "github.com/g0ulartleo/mirante/proto/sentinelruntime/v1"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRunner struct {
	signal signal.Signal
	err    error
}

func (f *fakeRunner) Check(ctx context.Context, alarmID, sentinelType string, config map[string]any) (signal.Signal, error) {
	return f.signal, f.err
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

func TestCheckAlarm_RuntimeInvalidConfigWritesUnknownAndSkipsRetry(t *testing.T) {
	alarmService := alarm.NewAlarmService(&fakeAlarmRepo{alarm: &alarm.Alarm{
		ID:   "alarm-1",
		Type: "endpoint-checker",
		Config: map[string]any{
			"url": "https://example.com",
		},
	}})
	signalRepo := &fakeSignalRepo{}
	signalService := signal.NewService(signalRepo)

	runtimeErr := &runtimeclient.RuntimeError{
		Code: runtimev1.RuntimeErrorCode_RUNTIME_ERROR_CODE_INVALID_CONFIG,
		Signal: signal.Signal{
			AlarmID:   "alarm-1",
			Status:    signal.StatusUnknown,
			Timestamp: time.Now(),
			Message:   "bad config",
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
	assert.Equal(t, "bad config", signalRepo.saved[0].Message)
}

func TestCheckAlarm_RuntimeCheckFailedWritesUnknownAndReturnsNil(t *testing.T) {
	alarmService := alarm.NewAlarmService(&fakeAlarmRepo{alarm: &alarm.Alarm{
		ID:   "alarm-1",
		Type: "endpoint-checker",
		Config: map[string]any{
			"url": "https://example.com",
		},
	}})
	signalRepo := &fakeSignalRepo{}
	signalService := signal.NewService(signalRepo)

	runtimeErr := &runtimeclient.RuntimeError{
		Code: runtimev1.RuntimeErrorCode_RUNTIME_ERROR_CODE_CHECK_FAILED,
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
		ID:   "alarm-1",
		Type: "endpoint-checker",
		Config: map[string]any{
			"url": "https://example.com",
		},
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
	assert.Contains(t, err.Error(), "failed to check sentinel via runner")
	assert.Empty(t, signalRepo.saved)
}
