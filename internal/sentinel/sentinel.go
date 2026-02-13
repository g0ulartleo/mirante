package sentinel

import (
	"context"

	"github.com/g0ulartleo/mirante/internal/signal"
)

type Sentinel interface {
	Check(ctx context.Context, alarmID string) (signal.Signal, error)
	Configure(config map[string]any) error
}
