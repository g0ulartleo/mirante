package client

import (
	"testing"
	"time"

	"github.com/g0ulartleo/mirante/internal/signal"
	runtimev1 "github.com/g0ulartleo/mirante/proto/sentinelruntime/v1"
	"github.com/stretchr/testify/assert"
)

func TestFromProtoStatus(t *testing.T) {
	assert.Equal(t, signal.StatusHealthy, fromProtoStatus(runtimev1.SignalStatus_SIGNAL_STATUS_HEALTHY))
	assert.Equal(t, signal.StatusUnhealthy, fromProtoStatus(runtimev1.SignalStatus_SIGNAL_STATUS_UNHEALTHY))
	assert.Equal(t, signal.StatusUnknown, fromProtoStatus(runtimev1.SignalStatus_SIGNAL_STATUS_UNKNOWN))
	assert.Equal(t, signal.StatusUnknown, fromProtoStatus(runtimev1.SignalStatus_SIGNAL_STATUS_UNSPECIFIED))
}

func TestFromResponseUsesNowWhenTimestampMissing(t *testing.T) {
	before := time.Now()
	sig := fromResponse("alarm-1", &runtimev1.CheckResponse{
		Status:  runtimev1.SignalStatus_SIGNAL_STATUS_UNHEALTHY,
		Message: "bad",
	})
	after := time.Now()

	assert.Equal(t, "alarm-1", sig.AlarmID)
	assert.Equal(t, signal.StatusUnhealthy, sig.Status)
	assert.Equal(t, "bad", sig.Message)
	assert.False(t, sig.Timestamp.Before(before))
	assert.False(t, sig.Timestamp.After(after))
}
