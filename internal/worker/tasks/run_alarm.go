package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/g0ulartleo/mirante/internal/alarm"
	runtimeclient "github.com/g0ulartleo/mirante/internal/alarm/runtime/client"
	"github.com/g0ulartleo/mirante/internal/signal"
	runtimev1 "github.com/g0ulartleo/mirante/proto/alarmruntime/v1"
	"github.com/hibiken/asynq"
)

const (
	TypeAlarmCheck = "alarm:check"
)

type AlarmCheckPayload struct {
	AlarmID string
}

type runAlarmRunner interface {
	RunAlarm(ctx context.Context, runtimeName string, alarmID string) (signal.Signal, error)
}

type taskEnqueuer interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
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
		asynq.Unique(30*time.Second),
	), nil
}

func HandleAlarmCheckTask(
	ctx context.Context,
	t *asynq.Task,
	runnerClient runAlarmRunner,
	signalService *signal.Service,
	alarmService *alarm.AlarmService,
	asyncClient taskEnqueuer,
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
	runnerClient runAlarmRunner,
	signalService *signal.Service,
	alarmService *alarm.AlarmService,
	asyncClient taskEnqueuer,
) error {
	alarmConfig, err := alarmService.GetAlarm(payload.AlarmID)
	if err != nil {
		return fmt.Errorf("failed to load alarm config: %v: %w", err, asynq.SkipRetry)
	}
	start := time.Now()
	log.Printf("Alarm check request alarm_id=%s", payload.AlarmID)
	sig, err := runnerClient.RunAlarm(ctx, alarmConfig.Runtime, payload.AlarmID)
	if err != nil {
		runtimeErr, ok := err.(*runtimeclient.RuntimeError)
		if !ok {
			log.Printf("Alarm check failed alarm_id=%s duration=%s error=%v", payload.AlarmID, time.Since(start), err)
			return fmt.Errorf("failed to check alarm via runtime: %w", err)
		}
		log.Printf("Alarm check response alarm_id=%s status=%s duration=%s runtime_error_code=%s message=%q", payload.AlarmID, runtimeErr.Signal.Status, time.Since(start), runtimeErr.Code.String(), runtimeErr.Signal.Message)

		if writeErr := signalService.WriteSignal(runtimeErr.Signal); writeErr != nil {
			return fmt.Errorf("failed to write signal: %w", writeErr)
		}

		switch runtimeErr.Code {
		case runtimev1.RuntimeErrorCode_RUNTIME_ERROR_CODE_UNSUPPORTED:
			return fmt.Errorf("alarm %q not supported by runtime %q: %v: %w", payload.AlarmID, alarmConfig.Runtime, runtimeErr, asynq.SkipRetry)
		default:
			return nil
		}
	}
	log.Printf("Alarm check response alarm_id=%s status=%s duration=%s message=%q", payload.AlarmID, sig.Status, time.Since(start), sig.Message)
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

	signals, err := signalService.GetAlarmLatestSignals(payload.AlarmID, 2)
	if err != nil {
		return fmt.Errorf("failed to get alarm latest signals: %w", err)
	}
	changed := len(signals) <= 1 || signals[0].Status != signals[1].Status
	if !changed {
		return nil
	}

	prevStatus := signal.StatusUnknown
	if len(signals) >= 2 {
		prevStatus = signals[1].Status
	}

	if alarmConfig.HasNotificationsEnabled() {
		task, err := NewAlarmNotifyTask(payload.AlarmID, sig, prevStatus)
		if err != nil {
			return fmt.Errorf("failed to create notify task: %w", err)
		}
		if _, err := asyncClient.Enqueue(task); err != nil {
			return fmt.Errorf("failed to enqueue task: %w", err)
		}
	}
	return nil
}
