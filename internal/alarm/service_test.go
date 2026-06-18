package alarm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeAlarmRepository struct {
	storedAlarm *Alarm
}

func (r *fakeAlarmRepository) Init() error {
	return nil
}

func (r *fakeAlarmRepository) GetAlarms() ([]*Alarm, error) {
	return nil, nil
}

func (r *fakeAlarmRepository) GetAlarm(alarmID string) (*Alarm, error) {
	return nil, nil
}

func (r *fakeAlarmRepository) SetAlarm(a *Alarm) error {
	r.storedAlarm = a
	return nil
}

func (r *fakeAlarmRepository) DeleteAlarm(alarmID string) error {
	return nil
}

func (r *fakeAlarmRepository) DeleteStaleAlarmsByRuntime(runtime string, keepIDs map[string]bool) error {
	return nil
}

func (r *fakeAlarmRepository) Close() error {
	return nil
}

func TestAlarmServiceSetAlarm(t *testing.T) {
	repo := &fakeAlarmRepository{}
	svc := NewAlarmService(repo)

	a := &Alarm{
		ID:       "test-alarm",
		Name:     "Test Alarm",
		Interval: "1m",
	}

	err := svc.SetAlarm(a)
	assert.NoError(t, err)
	assert.NotNil(t, repo.storedAlarm)
	assert.Equal(t, "test-alarm", repo.storedAlarm.ID)
}
