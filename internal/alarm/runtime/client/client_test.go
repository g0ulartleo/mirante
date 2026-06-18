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
			SlackWebhooks: []*runtimev1.SlackWebhookNotification{
				{Url: "https://hooks.slack.test/1"},
				{Url: "https://hooks.slack.test/2"},
			},
			Emails: []*runtimev1.EmailNotification{
				{To: []string{"a@example.com"}},
				{To: []string{"b@example.com", "c@example.com"}},
			},
			NotifyMissingSignals: true,
		},
	})

	assert.Len(t, a.Notifications.SlackWebhooks, 2)
	assert.Equal(t, "https://hooks.slack.test/1", a.Notifications.SlackWebhooks[0].URL)
	assert.Equal(t, "https://hooks.slack.test/2", a.Notifications.SlackWebhooks[1].URL)
	assert.Len(t, a.Notifications.Emails, 2)
	assert.Equal(t, []string{"a@example.com"}, a.Notifications.Emails[0].To)
	assert.Equal(t, []string{"b@example.com", "c@example.com"}, a.Notifications.Emails[1].To)
	assert.True(t, a.Notifications.NotifyMissingSignals)
}

func TestFromResponseConvertsDetails(t *testing.T) {
	object, err := structpb.NewStruct(map[string]any{"count": float64(3)})
	assert.NoError(t, err)

	sig := fromRunAlarmResponse("alarm-1", &runtimev1.RunAlarmResponse{
		Status: runtimev1.SignalStatus_SIGNAL_STATUS_WARNING,
		Details: []*runtimev1.RunAlarmDetail{
			{Title: "text", Value: &runtimev1.RunAlarmDetail_Text{Text: "hello"}},
			{Title: "object", Value: &runtimev1.RunAlarmDetail_Object{Object: object}},
			{Title: "table", Value: &runtimev1.RunAlarmDetail_Table{Table: &runtimev1.TableDetail{
				Columns: []string{"name", "count"},
				Rows: []*runtimev1.TableRow{
					{Cells: []string{"jobs", "3"}},
				},
			}}},
			{Title: "list", Value: &runtimev1.RunAlarmDetail_List{List: &runtimev1.ListDetail{Items: []string{"a", "b"}}}},
		},
	})

	assert.Len(t, sig.Details, 4)
	assert.Equal(t, signal.DetailTypeText, sig.Details[0].Type)
	assert.Equal(t, "hello", sig.Details[0].Text)
	assert.Equal(t, signal.DetailTypeObject, sig.Details[1].Type)
	assert.Equal(t, float64(3), sig.Details[1].Object["count"])
	assert.Equal(t, signal.DetailTypeTable, sig.Details[2].Type)
	assert.Equal(t, []string{"name", "count"}, sig.Details[2].Table.Columns)
	assert.Equal(t, [][]string{{"jobs", "3"}}, sig.Details[2].Table.Rows)
	assert.Equal(t, signal.DetailTypeList, sig.Details[3].Type)
	assert.Equal(t, []string{"a", "b"}, sig.Details[3].List)
}
