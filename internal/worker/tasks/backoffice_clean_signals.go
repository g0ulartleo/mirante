package tasks

import (
	"context"

	"github.com/g0ulartleo/mirante-alerts/internal/signal"
	"github.com/hibiken/asynq"
)

const (
	TypeBackofficeCleanSignals = "backoffice:clean-signals"
)

func NewBackofficeCleanSignalsTask() (*asynq.Task, error) {
	return asynq.NewTask(
		TypeBackofficeCleanSignals,
		nil,
		asynq.MaxRetry(3),
		asynq.TaskID(TypeBackofficeCleanSignals),
	), nil
}

func HandleBackofficeCleanSignalsTask(ctx context.Context, t *asynq.Task, signalService *signal.Service) error {
	return signalService.CleanOldSignals()
}
