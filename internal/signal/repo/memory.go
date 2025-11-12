package repo

import (
	"time"

	"github.com/g0ulartleo/mirante-alerts/internal/signal"
)

type MemorySignalRepository struct {
	signals map[string][]signal.Signal
}

func NewMemorySignalRepository() *MemorySignalRepository {
	return &MemorySignalRepository{signals: make(map[string][]signal.Signal)}
}

func (r *MemorySignalRepository) Init() error {
	r.signals = make(map[string][]signal.Signal)
	return nil
}

func (r *MemorySignalRepository) Save(signal signal.Signal) error {
	r.signals[signal.AlarmID] = append(r.signals[signal.AlarmID], signal)
	return nil
}

func (r *MemorySignalRepository) GetAlarmLatestSignals(alarmID string, limit int) ([]signal.Signal, error) {
	signals := r.signals[alarmID]
	if len(signals) == 0 {
		return nil, nil
	}
	return signals[len(signals)-limit:], nil
}

func (r *MemorySignalRepository) GetAlarmHealth(alarmID string) (signal.Status, error) {
	signals, err := r.GetAlarmLatestSignals(alarmID, 1)
	if err != nil {
		return signal.StatusUnknown, err
	}
	if len(signals) == 0 {
		return signal.StatusUnknown, nil
	}
	return signals[0].Status, nil
}

func (r *MemorySignalRepository) CleanOldSignals() error {
	return nil
}

func (r *MemorySignalRepository) Close() error {
	return nil
}

// naive in-memory scan
func (r *MemorySignalRepository) CountUnhealthySince(alarmID string, since time.Time) (int, error) {
	signals := r.signals[alarmID]
	count := 0
	for _, s := range signals {
		if s.Timestamp.After(since) || s.Timestamp.Equal(since) {
			if s.Status == signal.StatusUnhealthy {
				count++
			}
		}
	}
	return count, nil
}
