package sync

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/g0ulartleo/mirante/internal/alarm"
)

type alarmLister interface {
	ListAlarmsByRuntime(ctx context.Context) (map[string][]*alarm.Alarm, error)
}

type AlarmSyncer struct {
	lister alarmLister
	repo   alarm.AlarmRepository
}

type ValidationError struct {
	Err error
}

func (e *ValidationError) Error() string {
	return e.Err.Error()
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

type RuntimeListError struct {
	Err error
}

func (e *RuntimeListError) Error() string {
	return e.Err.Error()
}

func (e *RuntimeListError) Unwrap() error {
	return e.Err
}

func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

func New(lister alarmLister, repo alarm.AlarmRepository) *AlarmSyncer {
	return &AlarmSyncer{lister: lister, repo: repo}
}

func (s *AlarmSyncer) Sync(ctx context.Context) error {
	alarmsByRuntime, err := s.lister.ListAlarmsByRuntime(ctx)
	if err != nil {
		if len(alarmsByRuntime) == 0 {
			return &RuntimeListError{Err: fmt.Errorf("failed to list alarms from runtimes: %w", err)}
		}
		log.Printf("runtime sync continuing with partial alarm list: %v", err)
	}

	var alarms []*alarm.Alarm
	keepIDsByRuntime := map[string]map[string]bool{}
	for runtime, runtimeAlarms := range alarmsByRuntime {
		keepIDsByRuntime[runtime] = map[string]bool{}
		for _, a := range runtimeAlarms {
			a.Runtime = runtime
			keepIDsByRuntime[runtime][a.ID] = true
			alarms = append(alarms, a)
		}
	}

	if err := validateAlarms(alarms); err != nil {
		return &ValidationError{Err: err}
	}

	for _, a := range alarms {
		if err := s.repo.SetAlarm(a); err != nil {
			return fmt.Errorf("failed to save alarm %q: %w", a.ID, err)
		}
		log.Printf("synced alarm id=%s runtime=%s", a.ID, a.Runtime)
	}

	for runtime, keepIDs := range keepIDsByRuntime {
		if err := s.repo.DeleteStaleAlarmsByRuntime(runtime, keepIDs); err != nil {
			return fmt.Errorf("failed to delete stale alarms for runtime %q: %w", runtime, err)
		}
	}

	if err != nil {
		return &RuntimeListError{Err: err}
	}

	return nil
}

func validateAlarms(alarms []*alarm.Alarm) error {
	seen := map[string]string{}
	for _, a := range alarms {
		if a.ID == "" {
			return fmt.Errorf("alarm id is required (runtime=%q)", a.Runtime)
		}
		if a.Description == "" {
			return fmt.Errorf("alarm %q description is required", a.ID)
		}
		if a.Cron == "" && a.Interval == "" {
			return fmt.Errorf("alarm %q: cron or interval is required", a.ID)
		}
		if a.Cron != "" && a.Interval != "" {
			return fmt.Errorf("alarm %q: cron and interval cannot both be set", a.ID)
		}
		if a.Interval != "" {
			if _, err := time.ParseDuration(a.Interval); err != nil {
				return fmt.Errorf("alarm %q: invalid interval %q: %w", a.ID, a.Interval, err)
			}
		}
		if existing, ok := seen[a.ID]; ok {
			return fmt.Errorf("duplicate alarm id %q across runtimes %q and %q", a.ID, existing, a.Runtime)
		}
		seen[a.ID] = a.Runtime
	}
	return nil
}
