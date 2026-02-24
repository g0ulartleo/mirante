package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/g0ulartleo/mirante/internal/signal"
	runtimev1 "github.com/g0ulartleo/mirante/proto/sentinelruntime/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	conn    *grpc.ClientConn
	service runtimev1.SentinelRuntimeClient
	timeout time.Duration
}

type RuntimeError struct {
	Code   runtimev1.RuntimeErrorCode
	Signal signal.Signal
}

func (e *RuntimeError) Error() string {
	return e.Signal.Message
}

func New(addr string, timeout time.Duration) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sentinel runner: %w", err)
	}
	return &Client{
		conn:    conn,
		service: runtimev1.NewSentinelRuntimeClient(conn),
		timeout: timeout,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Check(ctx context.Context, alarmID, sentinelType string, config map[string]any) (signal.Signal, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return signal.Signal{}, fmt.Errorf("failed to marshal config: %w", err)
	}

	request := &runtimev1.CheckRequest{
		AlarmId:      alarmID,
		SentinelType: sentinelType,
		ConfigJson:   configJSON,
	}

	resp, err := c.checkWithRetry(ctx, request)
	if err != nil {
		return signal.Signal{}, err
	}

	sig := fromResponse(alarmID, resp)
	if resp.GetError() != nil {
		return sig, &RuntimeError{
			Code:   resp.GetError().GetCode(),
			Signal: sig,
		}
	}

	return sig, nil
}

func (c *Client) checkWithRetry(ctx context.Context, request *runtimev1.CheckRequest) (*runtimev1.CheckResponse, error) {
	const maxAttempts = 3
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		rpcCtx, cancel := context.WithTimeout(ctx, c.timeout)
		resp, err := c.service.Check(rpcCtx, request)
		cancel()
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if !isRetryableRPCError(err) || attempt == maxAttempts {
			break
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(attempt) * 200 * time.Millisecond):
		}
	}

	return nil, lastErr
}

func isRetryableRPCError(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	switch st.Code() {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted, codes.Aborted:
		return true
	default:
		return false
	}
}

func fromResponse(alarmID string, response *runtimev1.CheckResponse) signal.Signal {
	return signal.Signal{
		AlarmID:   alarmID,
		Status:    fromProtoStatus(response.GetStatus()),
		Timestamp: time.Now(),
		Message:   response.GetMessage(),
	}
}

func fromProtoStatus(status runtimev1.SignalStatus) signal.Status {
	switch status {
	case runtimev1.SignalStatus_SIGNAL_STATUS_HEALTHY:
		return signal.StatusHealthy
	case runtimev1.SignalStatus_SIGNAL_STATUS_UNHEALTHY:
		return signal.StatusUnhealthy
	default:
		return signal.StatusUnknown
	}
}
