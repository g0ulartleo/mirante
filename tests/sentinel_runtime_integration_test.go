package tests

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/g0ulartleo/mirante/internal/sentinel"
	runtimeclient "github.com/g0ulartleo/mirante/internal/sentinel/runtime/client"
	runtimeserver "github.com/g0ulartleo/mirante/internal/sentinel/runtime/server"
	"github.com/g0ulartleo/mirante/internal/signal"
	runtimev1 "github.com/g0ulartleo/mirante/proto/sentinelruntime/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type integrationSentinel struct{}

func (s *integrationSentinel) Configure(config map[string]any) error { return nil }

func (s *integrationSentinel) Check(ctx context.Context, alarmID string) (signal.Signal, error) {
	return signal.Signal{
		AlarmID:   alarmID,
		Status:    signal.StatusHealthy,
		Timestamp: time.Now(),
		Message:   "integration-ok",
	}, nil
}

func TestRuntimeClientCheckAgainstInMemoryGRPCServer(t *testing.T) {
	factory := sentinel.NewFactory()
	factory.Register("integration-sentinel", func() sentinel.Sentinel {
		return &integrationSentinel{}
	})

	grpcServer := grpc.NewServer()
	runtimev1.RegisterSentinelRuntimeServer(grpcServer, runtimeserver.New(factory))

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() {
		grpcServer.Stop()
		_ = lis.Close()
	})

	go func() {
		_ = grpcServer.Serve(lis)
	}()

	client, err := runtimeclient.New(lis.Addr().String(), 2*time.Second)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = client.Close()
	})

	sig, err := client.Check(
		context.Background(),
		"alarm-integration",
		"integration-sentinel",
		map[string]any{"foo": "bar"},
	)
	require.NoError(t, err)

	assert.Equal(t, signal.StatusHealthy, sig.Status)
	assert.Equal(t, "alarm-integration", sig.AlarmID)
	assert.Equal(t, "integration-ok", sig.Message)
}

func TestRuntimeClientCheckUnsupportedTypeReturnsRuntimeError(t *testing.T) {
	factory := sentinel.NewFactory()
	grpcServer := grpc.NewServer()
	runtimev1.RegisterSentinelRuntimeServer(grpcServer, runtimeserver.New(factory))

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() {
		grpcServer.Stop()
		_ = lis.Close()
	})

	go func() {
		_ = grpcServer.Serve(lis)
	}()

	client, err := runtimeclient.New(lis.Addr().String(), 2*time.Second)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = client.Close()
	})

	_, err = client.Check(
		context.Background(),
		"alarm-integration",
		"missing-sentinel",
		map[string]any{},
	)
	require.Error(t, err)

	runtimeErr, ok := err.(*runtimeclient.RuntimeError)
	require.True(t, ok)
	assert.Equal(t, runtimev1.RuntimeErrorCode_RUNTIME_ERROR_CODE_UNSUPPORTED_TYPE, runtimeErr.Code)
	assert.Equal(t, signal.StatusUnknown, runtimeErr.Signal.Status)
}
