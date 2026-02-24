package server

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/g0ulartleo/mirante/internal/sentinel"
	"github.com/g0ulartleo/mirante/internal/signal"
	runtimev1 "github.com/g0ulartleo/mirante/proto/sentinelruntime/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSentinel struct {
	configureErr error
	checkErr     error
	signal       signal.Signal
}

func (f *fakeSentinel) Configure(config map[string]any) error {
	return f.configureErr
}

func (f *fakeSentinel) Check(ctx context.Context, alarmID string) (signal.Signal, error) {
	return f.signal, f.checkErr
}

func TestCheckReturnsUnsupportedTypeError(t *testing.T) {
	factory := sentinel.NewFactory()
	srv := New(factory)

	resp, err := srv.Check(context.Background(), &runtimev1.CheckRequest{
		AlarmId:      "alarm-1",
		SentinelType: "does-not-exist",
		ConfigJson:   []byte(`{}`),
	})
	require.NoError(t, err)
	require.NotNil(t, resp.GetError())

	assert.Equal(t, runtimev1.SignalStatus_SIGNAL_STATUS_UNKNOWN, resp.GetStatus())
	assert.Equal(t, runtimev1.RuntimeErrorCode_RUNTIME_ERROR_CODE_UNSUPPORTED_TYPE, resp.GetError().GetCode())
}

func TestCheckReturnsInvalidConfigForMalformedJSON(t *testing.T) {
	factory := sentinel.NewFactory()
	factory.Register("fake", func() sentinel.Sentinel {
		return &fakeSentinel{}
	})
	srv := New(factory)

	resp, err := srv.Check(context.Background(), &runtimev1.CheckRequest{
		AlarmId:      "alarm-1",
		SentinelType: "fake",
		ConfigJson:   []byte(`{"invalid_json":`),
	})
	require.NoError(t, err)
	require.NotNil(t, resp.GetError())

	assert.Equal(t, runtimev1.RuntimeErrorCode_RUNTIME_ERROR_CODE_INVALID_CONFIG, resp.GetError().GetCode())
	assert.Equal(t, runtimev1.SignalStatus_SIGNAL_STATUS_UNKNOWN, resp.GetStatus())
}

func TestCheckReturnsSignalPayloadOnSuccess(t *testing.T) {
	expectedSignal := signal.Signal{
		AlarmID: "alarm-1",
		Status:  signal.StatusHealthy,
		Message: "all good",
	}

	factory := sentinel.NewFactory()
	factory.Register("fake", func() sentinel.Sentinel {
		return &fakeSentinel{signal: expectedSignal}
	})
	srv := New(factory)

	config, err := json.Marshal(map[string]any{"foo": "bar"})
	require.NoError(t, err)

	resp, err := srv.Check(context.Background(), &runtimev1.CheckRequest{
		AlarmId:      "alarm-1",
		SentinelType: "fake",
		ConfigJson:   config,
	})
	require.NoError(t, err)

	assert.Equal(t, runtimev1.SignalStatus_SIGNAL_STATUS_HEALTHY, resp.GetStatus())
	assert.Equal(t, expectedSignal.Message, resp.GetMessage())
	assert.Nil(t, resp.GetError())
}
