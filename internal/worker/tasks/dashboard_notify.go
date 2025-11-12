package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/g0ulartleo/mirante-alerts/internal/alarm"
	"github.com/g0ulartleo/mirante-alerts/internal/signal"
	"github.com/g0ulartleo/mirante-alerts/internal/web/dashboard"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

const (
	TypeDashboardNotify = "dashboard:notify"
)

type DashboardNotifyPayload struct {
	AlarmID string
	Signal  signal.Signal
}

func NewDashboardNotifyTask(alarmID string, signal signal.Signal) (*asynq.Task, error) {
	payload, err := json.Marshal(DashboardNotifyPayload{AlarmID: alarmID, Signal: signal})
	if err != nil {
		return nil, fmt.Errorf("json.Marshal failed: %w", err)
	}
	return asynq.NewTask(
		TypeDashboardNotify,
		payload,
		asynq.MaxRetry(1),
		asynq.TaskID(fmt.Sprintf("%s:%s", TypeDashboardNotify, alarmID)),
	), nil
}

func HandleDashboardNotifyTask(ctx context.Context, t *asynq.Task, signalService *signal.Service, alarmService *alarm.AlarmService, redisClient *redis.Client) error {
	var p DashboardNotifyPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	alarmsSignals, err := dashboard.GetAlarmSignals(signalService, alarmService)
	if err != nil {
		return fmt.Errorf("failed to get alarm signals: %w", err)
	}
	alarmsData, err := json.Marshal(alarmsSignals)
	if err != nil {
		return fmt.Errorf("failed to marshal alarms data: %w", err)
	}
	if err := redisClient.Publish(ctx, "dashboard:updates", alarmsData).Err(); err != nil {
		return fmt.Errorf("failed to publish update: %w", err)
	}

	return nil
}
