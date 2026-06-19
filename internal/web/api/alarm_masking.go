package api

import (
	"github.com/g0ulartleo/mirante/internal/alarm"
)

func MaskSensitiveData(a *alarm.Alarm) *alarm.Alarm {
	maskedAlarm := &alarm.Alarm{
		ID:            a.ID,
		Name:          a.Name,
		Description:   a.Description,
		HowToFix:      a.HowToFix,
		Runtime:       a.Runtime,
		Path:          make([]string, len(a.Path)),
		Cron:          a.Cron,
		Interval:      a.Interval,
		Notifications: a.Notifications,
	}
	copy(maskedAlarm.Path, a.Path)

	channels := make(map[string]alarm.NotificationChannel, len(a.Notifications.Channels))
	for key, ch := range a.Notifications.Channels {
		slackWebhooks := append([]alarm.SlackWebhookNotificationConfig(nil), ch.SlackWebhooks...)
		for i := range slackWebhooks {
			if slackWebhooks[i].URL != "" {
				slackWebhooks[i].URL = "****"
			}
		}
		emails := append([]alarm.EmailNotificationConfig(nil), ch.Emails...)
		channels[key] = alarm.NotificationChannel{
			SlackWebhooks: slackWebhooks,
			Emails:        emails,
		}
	}
	maskedAlarm.Notifications.Channels = channels
	return maskedAlarm
}
