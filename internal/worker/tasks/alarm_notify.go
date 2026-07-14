package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/notification"
	"github.com/g0ulartleo/mirante/internal/signal"
	"github.com/hibiken/asynq"
)

const (
	TypeAlarmNotify = "alarm:notify"
)

type AlarmNotifyPayload struct {
	AlarmID    string
	Signal     signal.Signal
	PrevStatus signal.Status
}

func NewAlarmNotifyTask(alarmID string, sig signal.Signal, prevStatus signal.Status) (*asynq.Task, error) {
	payload, err := json.Marshal(AlarmNotifyPayload{AlarmID: alarmID, Signal: sig, PrevStatus: prevStatus})
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
	log.Printf("alarm notify processing alarm_id=%s status=%s prev_status=%s", payload.AlarmID, payload.Signal.Status, payload.PrevStatus)
	errors := notification.Dispatch(alarmConfig, payload.Signal, payload.PrevStatus)
	if len(errors) > 0 {
		for _, err := range errors {
			log.Printf("alarm notify failed alarm_id=%s error=%v", payload.AlarmID, err)
		}
		return fmt.Errorf("failed to dispatch alarm notifications: %v", errors)
	}
	log.Printf("alarm notify sent alarm_id=%s status=%s prev_status=%s", payload.AlarmID, payload.Signal.Status, payload.PrevStatus)
	return nil
}
