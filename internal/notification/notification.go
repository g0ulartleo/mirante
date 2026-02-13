package notification

import (
	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/signal"
)

type Notification interface {
	Build(alarmConfig *alarm.Alarm, sig signal.Signal) error
	Send() error
}
