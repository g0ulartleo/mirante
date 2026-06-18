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
	maskedAlarm.Notifications.Emails = append([]alarm.EmailNotificationConfig(nil), a.Notifications.Emails...)
	maskedAlarm.Notifications.SlackWebhooks = append([]alarm.SlackWebhookNotificationConfig(nil), a.Notifications.SlackWebhooks...)
	for i := range maskedAlarm.Notifications.SlackWebhooks {
		if maskedAlarm.Notifications.SlackWebhooks[i].URL != "" {
			maskedAlarm.Notifications.SlackWebhooks[i].URL = "****"
		}
	}
	return maskedAlarm
}
