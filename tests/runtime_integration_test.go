package tests

import (
	"context"
	"net"
	"testing"
	"time"

	runtimeexample "github.com/g0ulartleo/mirante/internal/alarm/runtime/example"
	runtimeclient "github.com/g0ulartleo/mirante/internal/alarm/runtime/client"
	alarmsv1 "github.com/g0ulartleo/mirante/proto/alarms/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestRuntimeClientCheckReturnsUnsupportedForUnknownAlarm(t *testing.T) {
	grpcServer := grpc.NewServer()
	alarmsv1.RegisterAlarmRuntimeServer(grpcServer, runtimeexample.New())

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

	_, err = client.RunAlarm(context.Background(), "alarm-integration")
	require.Error(t, err)

	runtimeErr, ok := err.(*runtimeclient.RuntimeError)
	require.True(t, ok)
	assert.Equal(t, alarmsv1.RuntimeErrorCode_RUNTIME_ERROR_CODE_UNSUPPORTED, runtimeErr.Code)
}
