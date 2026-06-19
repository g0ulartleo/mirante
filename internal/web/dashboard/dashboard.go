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
		// Fetch a small trend window so clients (e.g. the TUI) can render
		// sparklines. Signals are returned oldest-first, latest last.
		signals, err := signalService.GetAlarmLatestSignals(a.ID, 12)
		if err != nil {
			log.Printf("Error fetching signals for alarm %s: %v", a.ID, err)
			signals = []signal.Signal{}
		}
		// Normalize to oldest-first ordering: backends disagree (SQL/redis
		// return newest-first, memory oldest-first). Consumers rely on the
		// latest signal being last (Signals[len-1]).
		sort.Slice(signals, func(i, j int) bool {
			return signals[i].Timestamp.Before(signals[j].Timestamp)
		})
		var lastChecked time.Time
		if len(signals) > 0 {
			lastChecked = signals[len(signals)-1].Timestamp
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
