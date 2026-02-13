package dashboard

import (
	"log"
	"sort"
	"time"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/signal"
)

func GetAlarmSignals(signalService *signal.Service, alarmService *alarm.AlarmService) ([]alarm.AlarmSignals, error) {
	alarmsSignals := make([]alarm.AlarmSignals, 0)
	alarms, err := alarmService.GetAlarms()
	if err != nil {
		return nil, err
	}
	sort.Slice(alarms, func(i, j int) bool { return alarms[i].Name < alarms[j].Name })

	for _, a := range alarms {
		signals, err := signalService.GetAlarmLatestSignals(a.ID, 1)
		if err != nil {
			log.Printf("Error fetching signals for alarm %s: %v", a.ID, err)
			signals = []signal.Signal{}
		}
		var lastChecked time.Time
		if len(signals) > 0 {
			lastChecked = signals[0].Timestamp
		}
		unhealthy24h, _ := signalService.CountUnhealthySince(a.ID, time.Now().Add(-24*time.Hour))
		alarmsSignals = append(alarmsSignals, alarm.AlarmSignals{
			Alarm:             *a,
			Signals:           signals,
			LastCheckedAt:     lastChecked,
			UnhealthyCount24h: unhealthy24h,
		})
	}
	return alarmsSignals, nil
}
