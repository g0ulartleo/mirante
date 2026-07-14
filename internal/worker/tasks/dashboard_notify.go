package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/signal"
	"github.com/g0ulartleo/mirante/internal/web/dashboard"
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
	log.Printf("dashboard notify processing alarm_id=%s status=%s", p.AlarmID, p.Signal.Status)

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
	log.Printf("dashboard notify sent alarm_id=%s status=%s", p.AlarmID, p.Signal.Status)

	return nil
}
