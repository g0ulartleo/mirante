package alarm

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func InitAlarms(alarmRepository AlarmRepository) error {
	alarms, err := getFileBasedAlarms()
	if err != nil {
		return fmt.Errorf("failed to get file based alarms: %w", err)
	}
	for _, alarm := range alarms {
		err := alarmRepository.SetAlarm(alarm)
		if err != nil {
			return fmt.Errorf("failed to save alarm: %w", err)
		}
	}
	return nil
}

func LoadAlarmConfig(path string) (*Alarm, error) {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read yml file: %w", err)
	}
	var alarm *Alarm
	err = yaml.Unmarshal(yamlFile, &alarm)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal yml file: %w", err)
	}
	err = ResolveAlarmConfigEnvVars(alarm)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve env vars in %s: %w", path, err)
	}
	if alarm.Interval == "" && alarm.Cron == "" {
		return nil, fmt.Errorf("misconfiguration for alarm %s: interval or cron is required", alarm.ID)
	}
	if alarm.Interval != "" && alarm.Cron != "" {
		return nil, fmt.Errorf("misconfiguration for alarm %s: interval and cron cannot both be set", alarm.ID)
	}
	if alarm.Interval != "" {
		interval, err := time.ParseDuration(alarm.Interval)
		if err != nil {
			return nil, fmt.Errorf("misconfiguration for alarm %s: failed to parse interval: %w", alarm.ID, err)
		}
		alarm.Cron = fmt.Sprintf("@every %s", interval)
	}
	if len(alarm.Path) == 0 {
		segments := strings.Split(path, "/")
		if len(segments) > 2 {
			alarm.Path = segments[2 : len(segments)-1]
		}
	}
	return alarm, nil
}

func getFileBasedAlarms() ([]*Alarm, error) {
	if _, err := os.Stat("config/alarms"); os.IsNotExist(err) {
		return nil, nil
	}
	alarms := make([]*Alarm, 0)
	err := filepath.Walk("config/alarms", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk alarms: %w", err)
		}
		if !info.IsDir() && strings.HasSuffix(path, ".yml") {
			config, err := LoadAlarmConfig(path)
			if err != nil {
				return fmt.Errorf("failed to load config from %s: %w", path, err)
			}
			log.Printf("loaded alarm id %s from path %s", config.ID, path)
			alarms = append(alarms, config)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk alarms: %w", err)
	}
	return alarms, nil
}
