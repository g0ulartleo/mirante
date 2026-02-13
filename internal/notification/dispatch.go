package notification

import (
	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/signal"
)

func Dispatch(alarmConfig *alarm.Alarm, sig signal.Signal) []error {
	notifications := []Notification{}
	if len(alarmConfig.Notifications.Email.To) > 0 {
		notifications = append(notifications, NewEmailNotification())
	}
	if alarmConfig.Notifications.Slack.WebhookURL != "" {
		notifications = append(notifications, NewSlackNotification())
	}

	errors := []error{}
	for _, n := range notifications {
		if err := n.Build(alarmConfig, sig); err != nil {
			errors = append(errors, err)
			continue
		}
		if err := n.Send(); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}
