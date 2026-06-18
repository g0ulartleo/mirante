package sync

import (
	"context"
	"errors"
	"testing"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeLister struct {
	alarmsByRuntime map[string][]*alarm.Alarm
	err             error
}

func (l *fakeLister) ListAlarmsByRuntime(ctx context.Context) (map[string][]*alarm.Alarm, error) {
	return l.alarmsByRuntime, l.err
}

type fakeRepo struct {
	alarms map[string]*alarm.Alarm
}

func newFakeRepo(initial ...*alarm.Alarm) *fakeRepo {
	repo := &fakeRepo{alarms: map[string]*alarm.Alarm{}}
	for _, a := range initial {
		repo.alarms[a.ID] = a
	}
	return repo
}

func (r *fakeRepo) Init() error { return nil }

func (r *fakeRepo) GetAlarms() ([]*alarm.Alarm, error) {
	alarms := make([]*alarm.Alarm, 0, len(r.alarms))
	for _, a := range r.alarms {
		alarms = append(alarms, a)
	}
	return alarms, nil
}

func (r *fakeRepo) GetAlarm(alarmID string) (*alarm.Alarm, error) {
	return r.alarms[alarmID], nil
}

func (r *fakeRepo) SetAlarm(a *alarm.Alarm) error {
	r.alarms[a.ID] = a
	return nil
}

func (r *fakeRepo) DeleteAlarm(alarmID string) error {
	delete(r.alarms, alarmID)
	return nil
}

func (r *fakeRepo) DeleteStaleAlarmsByRuntime(runtime string, keepIDs map[string]bool) error {
	for id, a := range r.alarms {
		if a.Runtime == runtime && !keepIDs[id] {
			delete(r.alarms, id)
		}
	}
	return nil
}

func (r *fakeRepo) Close() error { return nil }

func testAlarm(id string, runtime string) *alarm.Alarm {
	return &alarm.Alarm{
		ID:          id,
		Name:        id,
		Description: "description",
		Runtime:     runtime,
		Interval:    "1m",
	}
}

func TestSyncUpsertsAndDeletesOnlyStaleAlarms(t *testing.T) {
	repo := newFakeRepo(
		testAlarm("stale-go", "go"),
		testAlarm("existing-node", "node"),
	)
	lister := &fakeLister{alarmsByRuntime: map[string][]*alarm.Alarm{"go": {testAlarm("new-go", "go")}}}
	syncer := New(lister, repo)

	err := syncer.Sync(context.Background())
	require.NoError(t, err)

	assert.Contains(t, repo.alarms, "new-go")
	assert.NotContains(t, repo.alarms, "stale-go")
	assert.Contains(t, repo.alarms, "existing-node")
}

func TestSyncRejectsDuplicateIDsWithoutSaving(t *testing.T) {
	repo := newFakeRepo()
	lister := &fakeLister{alarmsByRuntime: map[string][]*alarm.Alarm{
		"go":   {testAlarm("dup", "go")},
		"node": {testAlarm("dup", "node")},
	}}
	syncer := New(lister, repo)

	err := syncer.Sync(context.Background())
	require.Error(t, err)
	assert.True(t, IsValidationError(err))
	assert.Empty(t, repo.alarms)
}

func TestSyncRejectsMissingDescriptionWithoutSaving(t *testing.T) {
	repo := newFakeRepo()
	a := testAlarm("missing-description", "go")
	a.Description = ""
	syncer := New(&fakeLister{alarmsByRuntime: map[string][]*alarm.Alarm{"go": {a}}}, repo)

	err := syncer.Sync(context.Background())
	require.Error(t, err)
	assert.True(t, IsValidationError(err))
	assert.Empty(t, repo.alarms)
}

func TestSyncRejectsInvalidIntervalWithoutSaving(t *testing.T) {
	repo := newFakeRepo()
	a := testAlarm("bad-interval", "go")
	a.Interval = "nope"
	syncer := New(&fakeLister{alarmsByRuntime: map[string][]*alarm.Alarm{"go": {a}}}, repo)

	err := syncer.Sync(context.Background())
	require.Error(t, err)
	assert.True(t, IsValidationError(err))
	assert.Empty(t, repo.alarms)
}

func TestSyncRuntimeListErrorKeepsCachedAlarms(t *testing.T) {
	repo := newFakeRepo(testAlarm("cached", "go"))
	syncer := New(&fakeLister{err: errors.New("runtime down")}, repo)

	err := syncer.Sync(context.Background())
	require.Error(t, err)
	var runtimeErr *RuntimeListError
	assert.True(t, errors.As(err, &runtimeErr))
	assert.Contains(t, repo.alarms, "cached")
}

func TestSyncPartialRuntimeListErrorStillSyncsSuccessfulRuntime(t *testing.T) {
	repo := newFakeRepo(
		testAlarm("stale-go", "go"),
		testAlarm("cached-node", "node"),
	)
	syncer := New(&fakeLister{
		alarmsByRuntime: map[string][]*alarm.Alarm{"go": {testAlarm("new-go", "go")}},
		err:             errors.New("node runtime down"),
	}, repo)

	err := syncer.Sync(context.Background())
	require.Error(t, err)
	var runtimeErr *RuntimeListError
	assert.True(t, errors.As(err, &runtimeErr))
	assert.Contains(t, repo.alarms, "new-go")
	assert.NotContains(t, repo.alarms, "stale-go")
	assert.Contains(t, repo.alarms, "cached-node")
}

func TestSyncSuccessfulEmptyRuntimeDeletesStaleAlarms(t *testing.T) {
	repo := newFakeRepo(
		testAlarm("stale-go", "go"),
		testAlarm("cached-node", "node"),
	)
	syncer := New(&fakeLister{alarmsByRuntime: map[string][]*alarm.Alarm{"go": {}}}, repo)

	err := syncer.Sync(context.Background())
	require.NoError(t, err)
	assert.NotContains(t, repo.alarms, "stale-go")
	assert.Contains(t, repo.alarms, "cached-node")
}
