package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/g0ulartleo/mirante/internal/alarm"
	runtimeclient "github.com/g0ulartleo/mirante/internal/sentinel/runtime/client"
	"github.com/g0ulartleo/mirante/internal/signal"
	runtimev1 "github.com/g0ulartleo/mirante/proto/sentinelruntime/v1"
	"github.com/hibiken/asynq"
)

const (
	TypeAlarmCheck = "alarm:check"
)

type AlarmCheckPayload struct {
	AlarmID string
}

type sentinelRunner interface {
	Check(ctx context.Context, alarmID, sentinelType string, config map[string]any) (signal.Signal, error)
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
	runnerClient sentinelRunner,
	signalService *signal.Service,
	alarmService *alarm.AlarmService,
	asyncClient *asynq.Client,
) error {
	var payload AlarmCheckPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	if err := checkAlarm(ctx, payload, runnerClient, signalService, alarmService, asyncClient); err != nil {
		return err
	}
	return nil
}

func checkAlarm(
	ctx context.Context,
	payload AlarmCheckPayload,
	runnerClient sentinelRunner,
	signalService *signal.Service,
	alarmService *alarm.AlarmService,
	asyncClient *asynq.Client,
) error {
	alarmConfig, err := alarmService.GetAlarm(payload.AlarmID)
	if err != nil {
		return fmt.Errorf("failed to load alarm config: %v: %w", err, asynq.SkipRetry)
	}
	start := time.Now()
	log.Printf("Sentinel check request alarm_id=%s sentinel_type=%s", payload.AlarmID, alarmConfig.Type)
	sig, err := runnerClient.Check(ctx, payload.AlarmID, alarmConfig.Type, alarmConfig.Config)
	if err != nil {
		runtimeErr, ok := err.(*runtimeclient.RuntimeError)
		if !ok {
			log.Printf("Sentinel check failed alarm_id=%s sentinel_type=%s duration=%s error=%v", payload.AlarmID, alarmConfig.Type, time.Since(start), err)
			return fmt.Errorf("failed to check sentinel via runner: %w", err)
		}
		log.Printf("Sentinel check response alarm_id=%s sentinel_type=%s status=%s duration=%s runtime_error_code=%s message=%q", payload.AlarmID, alarmConfig.Type, runtimeErr.Signal.Status, time.Since(start), runtimeErr.Code.String(), runtimeErr.Signal.Message)

		if writeErr := signalService.WriteSignal(runtimeErr.Signal); writeErr != nil {
			return fmt.Errorf("failed to write signal: %w", writeErr)
		}

		switch runtimeErr.Code {
		case runtimev1.RuntimeErrorCode_RUNTIME_ERROR_CODE_UNSUPPORTED_TYPE,
			runtimev1.RuntimeErrorCode_RUNTIME_ERROR_CODE_INVALID_CONFIG:
			return fmt.Errorf("failed to initialize sentinel via runner: %v: %w", runtimeErr, asynq.SkipRetry)
		default:
			return nil
		}
	}
	log.Printf("Sentinel check response alarm_id=%s sentinel_type=%s status=%s duration=%s message=%q", payload.AlarmID, alarmConfig.Type, sig.Status, time.Since(start), sig.Message)
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
