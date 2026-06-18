package notification

import (
	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/signal"
)

func Dispatch(alarmConfig *alarm.Alarm, sig signal.Signal) []error {
	notifications := []Notification{}
	for _, email := range alarmConfig.Notifications.Emails {
		if len(email.To) > 0 {
			notifications = append(notifications, NewEmailNotification(email.To))
		}
	}
	for _, webhook := range alarmConfig.Notifications.SlackWebhooks {
		if webhook.URL != "" {
			notifications = append(notifications, NewSlackNotification(webhook.URL))
		}
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
