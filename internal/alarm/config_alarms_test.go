package alarm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadAlarmConfig(t *testing.T) {
	tests := []struct {
		name          string
		yamlContent   string
		expectedAlarm *Alarm
		expectError   bool
	}{
		{
			name: "valid alarm with interval",
			yamlContent: `
id: test-alarm
name: Test Alarm
description: Test alarm configuration
type: endpoint-checker
interval: 30s
config:
  url: https://example.com
  expected_status: 200
notifications:
  email:
    to:
      - test@example.com
  slack:
    webhook_url: https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXX
`,
			expectedAlarm: &Alarm{
				ID:          "test-alarm",
				Name:        "Test Alarm",
				Description: "Test alarm configuration",
				Type:        "endpoint-checker",
				Interval:    "30s",
				Cron:        "@every 30s",
				Config: map[string]any{
					"url":             "https://example.com",
					"expected_status": int(200),
				},
				Notifications: AlarmNotifications{
					Email: EmailNotificationConfig{
						To: []string{"test@example.com"},
					},
					Slack: SlackNotificationConfig{
						WebhookURL: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXX",
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid alarm with cron",
			yamlContent: `
id: test-alarm
name: Test Alarm
description: Test alarm configuration
type: endpoint-checker
cron: "*/5 * * * *"
config:
  url: https://example.com
  expected_status: 200
notifications:
  email:
    to:
      - test@example.com
`,
			expectedAlarm: &Alarm{
				ID:          "test-alarm",
				Name:        "Test Alarm",
				Description: "Test alarm configuration",
				Type:        "endpoint-checker",
				Cron:        "*/5 * * * *",
				Config: map[string]any{
					"url":             "https://example.com",
					"expected_status": int(200),
				},
				Notifications: AlarmNotifications{
					Email: EmailNotificationConfig{
						To: []string{"test@example.com"},
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing required fields",
			yamlContent: `
name: Test Alarm
description: Test alarm configuration
type: endpoint-checker
config:
  url: https://example.com
`,
			expectError: true,
		},
		{
			name: "both interval and cron set",
			yamlContent: `
id: test-alarm
name: Test Alarm
description: Test alarm configuration
type: endpoint-checker
interval: 30s
cron: "*/5 * * * *"
config:
  url: https://example.com
`,
			expectError: true,
		},
		{
			name: "invalid interval format",
			yamlContent: `
id: test-alarm
name: Test Alarm
description: Test alarm configuration
type: endpoint-checker
interval: invalid
config:
  url: https://example.com
`,
			expectError: true,
		},
		{
			name: "alarm with path",
			yamlContent: `
id: test-alarm
name: Test Alarm
description: Test alarm configuration
type: endpoint-checker
interval: 30s
path: ["Project", "APIs"]
`,
			expectedAlarm: &Alarm{
				ID:          "test-alarm",
				Name:        "Test Alarm",
				Description: "Test alarm configuration",
				Type:        "endpoint-checker",
				Interval:    "30s",
				Cron:        "@every 30s",
				Path:        []string{"Project", "APIs"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := writeAlarmConfig(t, tt.yamlContent)
			alarm, err := LoadAlarmConfig(tmpFile)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedAlarm.ID, alarm.ID)
			assert.Equal(t, tt.expectedAlarm.Name, alarm.Name)
			assert.Equal(t, tt.expectedAlarm.Description, alarm.Description)
			assert.Equal(t, tt.expectedAlarm.Type, alarm.Type)
			assert.Equal(t, tt.expectedAlarm.Config, alarm.Config)
			assert.Equal(t, tt.expectedAlarm.Cron, alarm.Cron)
			assert.Equal(t, tt.expectedAlarm.Notifications, alarm.Notifications)
			if tt.expectedAlarm.Path != nil {
				assert.Equal(t, tt.expectedAlarm.Path, alarm.Path)
			}
		})
	}
}

func writeAlarmConfig(t *testing.T, yamlContent string) string {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yml")
	err := os.WriteFile(tmpFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	return tmpFile
}
