package notification

import (
	"os"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/signal"
)

func Dispatch(alarmConfig *alarm.Alarm, sig signal.Signal, prevStatus signal.Status) []error {
	if os.Getenv("IGNORE_NOTIFICATIONS") != "" {
		return nil
	}
	var errors []error

	if sig.Status == signal.StatusUnknown {
		for _, ch := range alarmConfig.Notifications.Channels {
			if ch.NotifyMissingSignals {
				errors = append(errors, dispatchCh(ch, alarmConfig, sig)...)
			}
		}
		return errors
	}

	key := channelKey(sig.Status, prevStatus)
	if key == "" {
		return nil
	}
	ch, ok := alarmConfig.Notifications.Channels[key]
	if !ok {
		return nil
	}
	return dispatchCh(ch, alarmConfig, sig)
}

func dispatchCh(ch alarm.NotificationChannel, alarmConfig *alarm.Alarm, sig signal.Signal) []error {
	notifications := []Notification{}
	for _, email := range ch.Emails {
		if len(email.To) > 0 {
			notifications = append(notifications, NewEmailNotification(email.To))
		}
	}
	for _, webhook := range ch.SlackWebhooks {
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

func channelKey(status, prevStatus signal.Status) string {
	switch status {
	case signal.StatusUnhealthy:
		return "critical"
	case signal.StatusWarning:
		return "warnings"
	case signal.StatusHealthy:
		switch prevStatus {
		case signal.StatusUnhealthy:
			return "critical"
		case signal.StatusWarning:
			return "warnings"
		}
	}
	return ""
}
