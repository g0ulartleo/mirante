package client

import (
	"testing"
	"time"

	"github.com/g0ulartleo/mirante/internal/signal"
	runtimev1 "github.com/g0ulartleo/mirante/proto/alarmruntime/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestFromProtoStatus(t *testing.T) {
	assert.Equal(t, signal.StatusHealthy, fromProtoStatus(runtimev1.SignalStatus_SIGNAL_STATUS_HEALTHY))
	assert.Equal(t, signal.StatusUnhealthy, fromProtoStatus(runtimev1.SignalStatus_SIGNAL_STATUS_UNHEALTHY))
	assert.Equal(t, signal.StatusWarning, fromProtoStatus(runtimev1.SignalStatus_SIGNAL_STATUS_WARNING))
	assert.Equal(t, signal.StatusUnknown, fromProtoStatus(runtimev1.SignalStatus_SIGNAL_STATUS_UNKNOWN))
	assert.Equal(t, signal.StatusUnknown, fromProtoStatus(runtimev1.SignalStatus_SIGNAL_STATUS_UNSPECIFIED))
}

func TestFromResponseUsesNowWhenTimestampMissing(t *testing.T) {
	before := time.Now()
	sig := fromRunAlarmResponse("alarm-1", &runtimev1.RunAlarmResponse{
		Status:  runtimev1.SignalStatus_SIGNAL_STATUS_UNHEALTHY,
		Message: "bad",
	})
	after := time.Now()

	assert.Equal(t, "alarm-1", sig.AlarmID)
	assert.Equal(t, signal.StatusUnhealthy, sig.Status)
	assert.Equal(t, "bad", sig.Message)
	assert.False(t, sig.Timestamp.Before(before))
	assert.False(t, sig.Timestamp.After(after))
}

func TestFromProtoAlarmKeepsRepeatedNotifications(t *testing.T) {
	a := fromProtoAlarm(&runtimev1.Alarm{
		Id:          "alarm-1",
		Name:        "Alarm 1",
		Description: "description",
		Notifications: &runtimev1.AlarmNotifications{
			Channels: map[string]*runtimev1.NotificationChannels{
				"critical": {
					SlackWebhooks: []*runtimev1.SlackWebhookNotification{
						{Url: "https://hooks.slack.test/1"},
					},
					NotifyMissingSignals: true,
				},
				"warnings": {
					SlackWebhooks: []*runtimev1.SlackWebhookNotification{
						{Url: "https://hooks.slack.test/2"},
					},
					Emails: []*runtimev1.EmailNotification{
						{To: []string{"b@example.com", "c@example.com"}},
					},
				},
			},
		},
	})

	assert.Len(t, a.Notifications.Channels, 2)
	assert.Len(t, a.Notifications.Channels["critical"].SlackWebhooks, 1)
	assert.Equal(t, "https://hooks.slack.test/1", a.Notifications.Channels["critical"].SlackWebhooks[0].URL)
	assert.True(t, a.Notifications.Channels["critical"].NotifyMissingSignals)
	assert.Len(t, a.Notifications.Channels["warnings"].SlackWebhooks, 1)
	assert.Equal(t, "https://hooks.slack.test/2", a.Notifications.Channels["warnings"].SlackWebhooks[0].URL)
	assert.Len(t, a.Notifications.Channels["warnings"].Emails, 1)
	assert.Equal(t, []string{"b@example.com", "c@example.com"}, a.Notifications.Channels["warnings"].Emails[0].To)
	assert.False(t, a.Notifications.Channels["warnings"].NotifyMissingSignals)
}

func TestFromResponseConvertsDetails(t *testing.T) {
	data, err := structpb.NewStruct(map[string]any{"count": float64(3)})
	assert.NoError(t, err)

	sig := fromRunAlarmResponse("alarm-1", &runtimev1.RunAlarmResponse{
		Status:  runtimev1.SignalStatus_SIGNAL_STATUS_WARNING,
		Details: []*structpb.Struct{data},
	})

	assert.Len(t, sig.Details, 1)
	assert.Equal(t, float64(3), sig.Details[0]["count"])
}
