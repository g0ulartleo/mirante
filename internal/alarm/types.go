package alarm

import (
	"time"

	"github.com/g0ulartleo/mirante/internal/signal"
)

type Alarm struct {
	ID            string             `yaml:"id"`
	Name          string             `yaml:"name"`
	Description   string             `yaml:"description"`
	HowToFix      string             `yaml:"how_to_fix"`
	Runtime       string             `yaml:"-"`
	Path          []string           `yaml:"path"`
	Cron          string             `yaml:"cron"`
	Interval      string             `yaml:"interval"`
	Notifications AlarmNotifications `yaml:"notifications"`
}

func (a *Alarm) HasNotificationsEnabled() bool {
	for _, ch := range a.Notifications.Channels {
		if len(ch.Emails) > 0 || len(ch.SlackWebhooks) > 0 {
			return true
		}
	}
	return false
}

type AlarmNotifications struct {
	Channels map[string]NotificationChannel `yaml:"channels"`
}

type NotificationChannel struct {
	Emails               []EmailNotificationConfig        `yaml:"emails"`
	SlackWebhooks        []SlackWebhookNotificationConfig `yaml:"slack_webhooks"`
	NotifyMissingSignals bool                              `yaml:"notify_missing_signals"`
}

type EmailNotificationConfig struct {
	To []string `yaml:"to"`
}

type SlackWebhookNotificationConfig struct {
	URL string `yaml:"url"`
}

type AlarmSignals struct {
	Alarm             Alarm
	Signals           []signal.Signal
	LastCheckedAt     time.Time
	UnhealthyCount24h int
}
