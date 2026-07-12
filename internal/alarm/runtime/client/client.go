package client

import (
	"context"
	"fmt"
	"time"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/signal"
	alarmsv1 "github.com/g0ulartleo/mirante/proto/alarms/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type Client struct {
	conn    *grpc.ClientConn
	service alarmsv1.AlarmRuntimeClient
	timeout time.Duration
}

type RuntimeError struct {
	Code   alarmsv1.RuntimeErrorCode
	Signal signal.Signal
}

func (e *RuntimeError) Error() string {
	return e.Signal.Message
}

func New(addr string, timeout time.Duration) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to alarm runtime: %w", err)
	}
	return &Client{
		conn:    conn,
		service: alarmsv1.NewAlarmRuntimeClient(conn),
		timeout: timeout,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) ListAlarms(ctx context.Context) ([]*alarm.Alarm, error) {
	rpcCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.service.ListAlarms(rpcCtx, &alarmsv1.ListAlarmsRequest{})
	if err != nil {
		return nil, err
	}

	alarms := make([]*alarm.Alarm, 0, len(resp.Alarms))
	for _, pa := range resp.Alarms {
		a := fromProtoAlarm(pa)
		alarms = append(alarms, &a)
	}
	return alarms, nil
}

func (c *Client) Health(ctx context.Context) error {
	rpcCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	_, err := c.service.Health(rpcCtx, &alarmsv1.HealthRequest{})
	return err
}

func fromProtoAlarm(pa *alarmsv1.Alarm) alarm.Alarm {
	a := alarm.Alarm{
		ID:          pa.GetId(),
		Name:        pa.GetName(),
		Description: pa.GetDescription(),
		HowToFix:    pa.GetHowToFix(),
		Path:        pa.GetPath(),
		Cron:        pa.GetCron(),
		Interval:    pa.GetInterval(),
	}
	channels := map[string]alarm.NotificationChannel{}
	for k, ch := range pa.GetNotifications().GetChannels() {
		nc := alarm.NotificationChannel{
			NotifyMissingSignals: ch.GetNotifyMissingSignals(),
		}
		for _, sn := range ch.GetSlackWebhooks() {
			nc.SlackWebhooks = append(nc.SlackWebhooks, alarm.SlackWebhookNotificationConfig{URL: sn.GetUrl()})
		}
		for _, en := range ch.GetEmails() {
			nc.Emails = append(nc.Emails, alarm.EmailNotificationConfig{To: en.GetTo()})
		}
		channels[k] = nc
	}
	a.Notifications.Channels = channels
	return a
}

func (c *Client) RunAlarm(ctx context.Context, alarmID string) (signal.Signal, error) {
	request := &alarmsv1.RunAlarmRequest{
		AlarmId: alarmID,
	}

	resp, err := c.runAlarmWithRetry(ctx, request)
	if err != nil {
		if isRetryableRPCError(err) {
			sig := signal.Signal{
				AlarmID:   alarmID,
				Status:    signal.StatusUnknown,
				Timestamp: time.Now(),
				Message:   "runtime unreachable: " + err.Error(),
			}
			return sig, &RuntimeError{
				Code:   alarmsv1.RuntimeErrorCode_RUNTIME_ERROR_CODE_INTERNAL,
				Signal: sig,
			}
		}
		return signal.Signal{}, err
	}

	sig := fromRunAlarmResponse(alarmID, resp)
	if resp.GetError() != nil {
		return sig, &RuntimeError{
			Code:   resp.GetError().GetCode(),
			Signal: sig,
		}
	}

	return sig, nil
}

func (c *Client) runAlarmWithRetry(ctx context.Context, request *alarmsv1.RunAlarmRequest) (*alarmsv1.RunAlarmResponse, error) {
	const maxAttempts = 3
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		rpcCtx, cancel := context.WithTimeout(ctx, c.timeout)
		resp, err := c.service.RunAlarm(rpcCtx, request)
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

func fromRunAlarmResponse(alarmID string, response *alarmsv1.RunAlarmResponse) signal.Signal {
	return signal.Signal{
		AlarmID:   alarmID,
		Status:    fromProtoStatus(response.GetStatus()),
		Timestamp: time.Now(),
		Message:   response.GetMessage(),
		Details:   fromProtoDetails(response.GetDetails()),
	}
}

func fromProtoDetails(details []*structpb.Struct) []map[string]any {
	converted := make([]map[string]any, 0, len(details))
	for _, detail := range details {
		if detail == nil {
			continue
		}
		converted = append(converted, detail.AsMap())
	}
	return converted
}

func fromProtoStatus(status alarmsv1.SignalStatus) signal.Status {
	switch status {
	case alarmsv1.SignalStatus_SIGNAL_STATUS_HEALTHY:
		return signal.StatusHealthy
	case alarmsv1.SignalStatus_SIGNAL_STATUS_UNHEALTHY:
		return signal.StatusUnhealthy
	case alarmsv1.SignalStatus_SIGNAL_STATUS_WARNING:
		return signal.StatusWarning
	default:
		return signal.StatusUnknown
	}
}
