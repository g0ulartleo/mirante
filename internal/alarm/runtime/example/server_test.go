package example

import (
	"context"
	"testing"

	alarmsv1 "github.com/g0ulartleo/mirante/proto/alarms/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunAlarmReturnsUnsupportedWhenNoAlarmsLoaded(t *testing.T) {
	srv := New()

	resp, err := srv.RunAlarm(context.Background(), &alarmsv1.RunAlarmRequest{
		AlarmId: "alarm-1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp.GetError())

	assert.Equal(t, alarmsv1.SignalStatus_SIGNAL_STATUS_UNKNOWN, resp.GetStatus())
	assert.Equal(t, alarmsv1.RuntimeErrorCode_RUNTIME_ERROR_CODE_UNSUPPORTED, resp.GetError().GetCode())
	assert.Contains(t, resp.GetError().GetMessage(), "not found")
}

func TestRunAlarmRejectsEmptyAlarmID(t *testing.T) {
	srv := New()

	_, err := srv.RunAlarm(context.Background(), &alarmsv1.RunAlarmRequest{})
	require.Error(t, err)
}

func TestRunAlarmRejectsNilRequest(t *testing.T) {
	srv := New()

	_, err := srv.RunAlarm(context.Background(), nil)
	require.Error(t, err)
}
