package repo

import (
	"github.com/g0ulartleo/mirante/internal/alarm"
)

func New() (alarm.AlarmRepository, error) {
	return NewRedisAlarmRepository()
}
