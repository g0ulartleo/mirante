package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/signal"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAPIAlarmRepo struct{}

func (r *fakeAPIAlarmRepo) Init() error                                   { return nil }
func (r *fakeAPIAlarmRepo) GetAlarms() ([]*alarm.Alarm, error)            { return nil, nil }
func (r *fakeAPIAlarmRepo) GetAlarm(alarmID string) (*alarm.Alarm, error) { return nil, nil }
func (r *fakeAPIAlarmRepo) SetAlarm(alarm *alarm.Alarm) error             { return nil }
func (r *fakeAPIAlarmRepo) DeleteAlarm(alarmID string) error              { return nil }
func (r *fakeAPIAlarmRepo) DeleteStaleAlarmsByRuntime(runtime string, keepIDs map[string]bool) error {
	return nil
}
func (r *fakeAPIAlarmRepo) Close() error { return nil }

type fakeAPISignalRepo struct{}

func (r *fakeAPISignalRepo) Init() error              { return nil }
func (r *fakeAPISignalRepo) Close() error             { return nil }
func (r *fakeAPISignalRepo) Save(signal.Signal) error { return nil }
func (r *fakeAPISignalRepo) GetAlarmLatestSignals(alarmID string, limit int) ([]signal.Signal, error) {
	return nil, nil
}
func (r *fakeAPISignalRepo) GetAlarmSignalsSince(alarmID string, since time.Time) ([]signal.Signal, error) {
	return nil, nil
}
func (r *fakeAPISignalRepo) GetAlarmHealth(alarmID string) (signal.Status, error) {
	return signal.StatusUnknown, nil
}
func (r *fakeAPISignalRepo) CountUnhealthySince(alarmID string, since time.Time) (int, error) {
	return 0, nil
}
func (r *fakeAPISignalRepo) CleanOldSignals() error { return nil }

func TestAlarmMutationEndpointsAreUnavailable(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "mirante.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte("auth:\n  api_key: \"\"\n"), 0644))
	t.Setenv("MIRANTE_CONFIG", configPath)

	e := echo.New()
	RegisterRoutes(
		e,
		signal.NewService(&fakeAPISignalRepo{}),
		alarm.NewAlarmService(&fakeAPIAlarmRepo{}),
		nil,
		nil,
	)

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/alarms/alarm-1", nil)
	deleteRec := httptest.NewRecorder()
	e.ServeHTTP(deleteRec, deleteReq)
	assert.Equal(t, http.StatusNotFound, deleteRec.Code)

	postReq := httptest.NewRequest(http.MethodPost, "/api/alarms", nil)
	postRec := httptest.NewRecorder()
	e.ServeHTTP(postRec, postReq)
	assert.Equal(t, http.StatusNotFound, postRec.Code)
}

func TestMaskSensitiveDataMasksAllSlackWebhookURLs(t *testing.T) {
	a := &alarm.Alarm{
		ID: "alarm-1",
		Notifications: alarm.AlarmNotifications{
			Channels: map[string]alarm.NotificationChannel{
				"critical": {
					SlackWebhooks: []alarm.SlackWebhookNotificationConfig{
						{URL: "https://hooks.slack.test/1"},
						{URL: "https://hooks.slack.test/2"},
					},
				},
			},
		},
	}

	masked := MaskSensitiveData(a)

	require.Len(t, masked.Notifications.Channels["critical"].SlackWebhooks, 2)
	assert.Equal(t, "****", masked.Notifications.Channels["critical"].SlackWebhooks[0].URL)
	assert.Equal(t, "****", masked.Notifications.Channels["critical"].SlackWebhooks[1].URL)
	assert.Equal(t, "https://hooks.slack.test/1", a.Notifications.Channels["critical"].SlackWebhooks[0].URL)
}
