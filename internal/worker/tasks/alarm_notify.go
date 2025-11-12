package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/g0ulartleo/mirante-alerts/internal/alarm"
	"github.com/g0ulartleo/mirante-alerts/internal/notification"
	"github.com/g0ulartleo/mirante-alerts/internal/signal"
	"github.com/hibiken/asynq"
)

const (
	TypeAlarmNotify = "alarm:notify"
)

type AlarmNotifyPayload struct {
	AlarmID string
	Signal  signal.Signal
}

func NewAlarmNotifyTask(alarmID string, sig signal.Signal) (*asynq.Task, error) {
	payload, err := json.Marshal(AlarmNotifyPayload{AlarmID: alarmID, Signal: sig})
	if err != nil {
		return nil, fmt.Errorf("json.Marshal failed: %w", err)
	}
	return asynq.NewTask(
		TypeAlarmNotify,
		payload,
		asynq.MaxRetry(1),
		asynq.TaskID(fmt.Sprintf("%s:%s", TypeAlarmNotify, alarmID)),
	), nil
}

func HandleAlarmNotifyTask(ctx context.Context, t *asynq.Task, alarmService *alarm.AlarmService) error {
	var payload AlarmNotifyPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %w", err)
	}
	alarmConfig, err := alarmService.GetAlarm(payload.AlarmID)
	if err != nil {
		return fmt.Errorf("failed to get alarm config: %w", err)
	}
	errors := notification.Dispatch(alarmConfig, payload.Signal)
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Println("error dispatching alarm notification: %w", err)
		}
		return fmt.Errorf("failed to dispatch alarm notifications: %v", errors)
	}
	return nil
}
