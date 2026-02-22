package alarm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func (r *fakeAlarmRepository) Close() error {
	return nil
}

func TestAlarmServiceSetAlarm_ResolvesEnvVarsOnInsert(t *testing.T) {
	t.Setenv("SERVICE_URL", "https://api.example.com")

	repo := &fakeAlarmRepository{}
	svc := NewAlarmService(repo)

	a := &Alarm{
		ID:       "service-env",
		Name:     "Service Env",
		Type:     "endpoint-checker",
		Interval: "1m",
		Config: map[string]any{
			"url": "${SERVICE_URL}",
		},
	}

	err := svc.SetAlarm(a)
	require.NoError(t, err)
	require.NotNil(t, repo.storedAlarm)
	assert.Equal(t, "https://api.example.com", repo.storedAlarm.Config["url"])
}

func TestAlarmServiceSetAlarm_FailsWhenEnvVarMissing(t *testing.T) {
	repo := &fakeAlarmRepository{}
	svc := NewAlarmService(repo)

	a := &Alarm{
		ID:       "service-missing-env",
		Name:     "Service Missing Env",
		Type:     "endpoint-checker",
		Interval: "1m",
		Config: map[string]any{
			"url": "${MISSING_SERVICE_URL}",
		},
	}

	err := svc.SetAlarm(a)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve env vars for alarm service-missing-env")
	assert.Contains(t, err.Error(), "missing environment variable: MISSING_SERVICE_URL")
	assert.Nil(t, repo.storedAlarm)
}
