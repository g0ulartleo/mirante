package api

import (
	"strings"

	"github.com/g0ulartleo/mirante/internal/alarm"
)

func MaskSensitiveData(a *alarm.Alarm) *alarm.Alarm {
	maskedAlarm := &alarm.Alarm{
		ID:            a.ID,
		Name:          a.Name,
		Description:   a.Description,
		Path:          make([]string, len(a.Path)),
		Type:          a.Type,
		Config:        make(map[string]any),
		Cron:          a.Cron,
		Interval:      a.Interval,
		Notifications: a.Notifications,
	}
	copy(maskedAlarm.Path, a.Path)
	for key, value := range a.Config {
		maskedAlarm.Config[key] = maskConfigValue(key, value)
	}
	return maskedAlarm
}

func maskConfigValue(key string, value any) any {
	keyLower := strings.ToLower(key)

	if strings.Contains(keyLower, "password") ||
		strings.Contains(keyLower, "token") ||
		strings.Contains(keyLower, "secret") ||
		strings.Contains(keyLower, "key") {
		return "****"
	}

	if nestedMap, ok := value.(map[string]any); ok {
		maskedMap := make(map[string]any)
		for nestedKey, nestedValue := range nestedMap {
			maskedMap[nestedKey] = maskConfigValue(nestedKey, nestedValue)
		}
		return maskedMap
	}

	return value
}
