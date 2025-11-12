package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/g0ulartleo/mirante-alerts/internal/signal"
	"github.com/hibiken/asynq"
)

const (
	TypeSignalWrite = "signal:write"
)

type SignalWritePayload struct {
	Signal signal.Signal
}

func NewSignalWriteTask(signal signal.Signal) (*asynq.Task, error) {
	payload, err := json.Marshal(SignalWritePayload{Signal: signal})
	if err != nil {
		return nil, fmt.Errorf("json.Marshal failed: %v", err)
	}
	return asynq.NewTask(
		TypeSignalWrite,
		payload,
		asynq.MaxRetry(3),
		asynq.TaskID(fmt.Sprintf("%s:%s", TypeSignalWrite, signal.AlarmID)),
	), nil
}

func HandleSignalWriteTask(ctx context.Context, t *asynq.Task, signalService *signal.Service) error {
	var p SignalWritePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("Writing signal: signal=%v", p.Signal)
	return signalService.WriteSignal(p.Signal)
}
