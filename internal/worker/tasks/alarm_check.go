package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/sentinel"
	"github.com/g0ulartleo/mirante/internal/signal"
	"github.com/hibiken/asynq"
)

const (
	TypeAlarmCheck = "alarm:check"
)

type AlarmCheckPayload struct {
	AlarmID string
}

func NewAlarmCheckTask(alarmID string) (*asynq.Task, error) {
	payload, err := json.Marshal(AlarmCheckPayload{AlarmID: alarmID})
	if err != nil {
		return nil, fmt.Errorf("json.Marshal failed: %v", err)
	}
	return asynq.NewTask(
		TypeAlarmCheck,
		payload,
		asynq.MaxRetry(1),
		asynq.TaskID(fmt.Sprintf("%s:%s", TypeAlarmCheck, alarmID)),
	), nil
}

func HandleAlarmCheckTask(
	ctx context.Context,
	t *asynq.Task,
	sentinelFactory *sentinel.SentinelFactory,
	signalService *signal.Service,
	alarmService *alarm.AlarmService,
	asyncClient *asynq.Client,
) error {
	var payload AlarmCheckPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	if err := checkAlarm(ctx, payload, sentinelFactory, signalService, alarmService, asyncClient); err != nil {
		return err
	}
	return nil
}

func checkAlarm(
	ctx context.Context,
	payload AlarmCheckPayload,
	sentinelFactory *sentinel.SentinelFactory,
	signalService *signal.Service,
	alarmService *alarm.AlarmService,
	asyncClient *asynq.Client,
) error {
	alarmConfig, err := alarmService.GetAlarm(payload.AlarmID)
	if err != nil {
		return fmt.Errorf("failed to load alarm config: %v: %w", err, asynq.SkipRetry)
	}
	sentinel, err := initializeSentinel(alarmConfig, sentinelFactory)
	if err != nil {
		writeErr := signalService.WriteSignal(signal.Signal{
			AlarmID:   payload.AlarmID,
			Status:    signal.StatusUnknown,
			Timestamp: time.Now(),
			Message:   fmt.Sprintf("failed to initialize sentinel: %v", err),
		})
		if writeErr != nil {
			return fmt.Errorf("failed to write signal: %w", writeErr)
		}
		return err
	}
	log.Printf("Sentinel checking alarm ID %s", payload.AlarmID)
	sig, err := sentinel.Check(ctx, payload.AlarmID)
	if err != nil {
		err = signalService.WriteSignal(signal.Signal{
			AlarmID:   payload.AlarmID,
			Status:    signal.StatusUnknown,
			Timestamp: time.Now(),
			Message:   fmt.Sprintf("failed to check sentinel: %v", err),
		})
		if err != nil {
			return fmt.Errorf("failed to write signal: %w", err)
		}
		return nil
	}
	log.Printf("Alarm %s returned signal: %v", payload.AlarmID, sig)
	err = signalService.WriteSignal(sig)
	if err != nil {
		return fmt.Errorf("failed to write signal: %w", err)
	}
	dashboardTask, err := NewDashboardNotifyTask(payload.AlarmID, sig)
	if err != nil {
		return fmt.Errorf("failed to create dashboard notify task: %w", err)
	}
	if _, err := asyncClient.Enqueue(dashboardTask); err != nil {
		return fmt.Errorf("failed to enqueue dashboard notify task: %w", err)
	}

	changed, err := signalService.AlarmHasChangedStatus(payload.AlarmID)
	if err != nil {
		return fmt.Errorf("failed to get alarm latest signals: %w", err)
	}
	if !changed {
		return nil
	}

	if alarmConfig.HasNotificationsEnabled() {
		if sig.Status == signal.StatusUnknown && !alarmConfig.Notifications.NotifyMissingSignals {
			return nil
		}
		task, err := NewAlarmNotifyTask(payload.AlarmID, sig)
		if err != nil {
			return fmt.Errorf("failed to create notify task: %w", err)
		}
		if _, err := asyncClient.Enqueue(task); err != nil {
			return fmt.Errorf("failed to enqueue task: %w", err)
		}
	}
	return nil
}

func initializeSentinel(alarmConfig *alarm.Alarm, sentinelFactory *sentinel.SentinelFactory) (sentinel.Sentinel, error) {
	sentinel, err := sentinelFactory.Create(alarmConfig.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get sentinel from factory: %v: %w", err, asynq.SkipRetry)
	}
	err = sentinel.Configure(alarmConfig.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to configure sentinel: %v: %w", err, asynq.SkipRetry)
	}
	return sentinel, nil
}
