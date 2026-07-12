package client

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/g0ulartleo/mirante/internal/config"
	"github.com/g0ulartleo/mirante/internal/signal"
	alarmsv1 "github.com/g0ulartleo/mirante/proto/alarms/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type testRuntimeServer struct {
	alarmsv1.UnimplementedAlarmRuntimeServer
	status alarmsv1.SignalStatus
	calls  atomic.Int32
}

func (s *testRuntimeServer) RunAlarm(ctx context.Context, req *alarmsv1.RunAlarmRequest) (*alarmsv1.RunAlarmResponse, error) {
	s.calls.Add(1)
	return &alarmsv1.RunAlarmResponse{
		Status:  s.status,
		Message: s.status.String(),
	}, nil
}

func (s *testRuntimeServer) ListAlarms(ctx context.Context, req *alarmsv1.ListAlarmsRequest) (*alarmsv1.ListAlarmsResponse, error) {
	return &alarmsv1.ListAlarmsResponse{}, nil
}

func TestRouterCheckSelectsExactRuntime(t *testing.T) {
	runtimeA := &testRuntimeServer{status: alarmsv1.SignalStatus_SIGNAL_STATUS_HEALTHY}
	runtimeB := &testRuntimeServer{status: alarmsv1.SignalStatus_SIGNAL_STATUS_WARNING}
	addrA := startRuntimeServer(t, runtimeA)
	addrB := startRuntimeServer(t, runtimeB)

	router, err := NewRouter(config.AlarmRuntimeConfig{
		Timeout: "2s",
		Runtimes: map[string]config.RuntimeConfig{
			"a": {Addr: addrA},
			"b": {Addr: addrB},
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, router.Close()) })

	sig, err := router.RunAlarm(context.Background(), "b", "alarm-1")
	require.NoError(t, err)

	assert.Equal(t, signal.StatusWarning, sig.Status)
	assert.Equal(t, int32(0), runtimeA.calls.Load())
	assert.Equal(t, int32(1), runtimeB.calls.Load())
}

func startRuntimeServer(t *testing.T, srv alarmsv1.AlarmRuntimeServer) string {
	t.Helper()
	grpcServer := grpc.NewServer()
	alarmsv1.RegisterAlarmRuntimeServer(grpcServer, srv)
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() {
		grpcServer.Stop()
		_ = lis.Close()
	})
	go func() {
		_ = grpcServer.Serve(lis)
	}()
	time.Sleep(10 * time.Millisecond)
	return lis.Addr().String()
}
