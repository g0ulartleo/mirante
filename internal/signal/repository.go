package signal

import "time"

type SignalRepository interface {
	Init() error
	Close() error
	Save(signal Signal) error
	GetAlarmLatestSignals(alarmID string, limit int) ([]Signal, error)
	GetAlarmSignalsSince(alarmID string, since time.Time) ([]Signal, error)
	GetAlarmHealth(alarmID string) (Status, error)
	CountUnhealthySince(alarmID string, since time.Time) (int, error)
	CleanOldSignals() error
}
