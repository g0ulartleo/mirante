package notification

import (
	"log"
	"os"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/signal"
)

func Dispatch(alarmConfig *alarm.Alarm, sig signal.Signal, prevStatus signal.Status) []error {
	if os.Getenv("IGNORE_NOTIFICATIONS") != "" {
		log.Printf("notification skipped alarm_id=%s status=%s reason=ignore_notifications", alarmConfig.ID, sig.Status)
		return nil
	}
	var errors []error

	if sig.Status == signal.StatusUnknown {
		for key, ch := range alarmConfig.Notifications.Channels {
			if ch.NotifyMissingSignals {
				log.Printf("notification channel selected alarm_id=%s status=%s prev_status=%s channel=%s", alarmConfig.ID, sig.Status, prevStatus, key)
				errors = append(errors, dispatchCh(key, ch, alarmConfig, sig)...)
			}
		}
		return errors
	}

	key := channelKey(sig.Status, prevStatus)
	if key == "" {
		log.Printf("notification skipped alarm_id=%s status=%s prev_status=%s reason=no_channel_key", alarmConfig.ID, sig.Status, prevStatus)
		return nil
	}
	ch, ok := alarmConfig.Notifications.Channels[key]
	if !ok {
		log.Printf("notification skipped alarm_id=%s status=%s prev_status=%s channel=%s reason=channel_missing", alarmConfig.ID, sig.Status, prevStatus, key)
		return nil
	}
	log.Printf("notification channel selected alarm_id=%s status=%s prev_status=%s channel=%s", alarmConfig.ID, sig.Status, prevStatus, key)
	return dispatchCh(key, ch, alarmConfig, sig)
}

func dispatchCh(key string, ch alarm.NotificationChannel, alarmConfig *alarm.Alarm, sig signal.Signal) []error {
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
	log.Printf("notification targets alarm_id=%s channel=%s slack_webhooks=%d emails=%d", alarmConfig.ID, key, len(ch.SlackWebhooks), len(ch.Emails))
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
