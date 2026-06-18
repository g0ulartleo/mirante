package templates

import (
	"strings"
	"testing"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/signal"
)

func makeFixtures() []alarm.AlarmSignals {
	return []alarm.AlarmSignals{
		{Alarm: alarm.Alarm{Name: "A Env1", Path: []string{"GroupA", "Env1"}}},
		{Alarm: alarm.Alarm{Name: "A Env2", Path: []string{"GroupA", "Env2"}}},
		{Alarm: alarm.Alarm{Name: "Metric One", Path: []string{"GroupB"}}},
		{Alarm: alarm.Alarm{Name: "Metric Two", Path: []string{"GroupB"}}},
	}
}

func TestWarningStatusUsesWarningColors(t *testing.T) {
	alarmSignals := alarm.AlarmSignals{Signals: []signal.Signal{{Status: signal.StatusWarning}}}

	if got := getAlarmStatusColor(alarmSignals); got != "bg-yellow-500" {
		t.Fatalf("getAlarmStatusColor warning = %s; want bg-yellow-500", got)
	}
	if got := getAlarmStatusDotColor(alarmSignals); got != "bg-yellow-300" {
		t.Fatalf("getAlarmStatusDotColor warning = %s; want bg-yellow-300", got)
	}
	if got := getGroupStatus([]alarm.AlarmSignals{alarmSignals}); got != "bg-yellow-500" {
		t.Fatalf("getGroupStatus warning = %s; want bg-yellow-500", got)
	}
}

func TestUnhealthyOverridesWarningGroupStatus(t *testing.T) {
	got := getGroupStatus([]alarm.AlarmSignals{
		{Signals: []signal.Signal{{Status: signal.StatusWarning}}},
		{Signals: []signal.Signal{{Status: signal.StatusUnhealthy}}},
	})
	if got != "bg-red-500" {
		t.Fatalf("getGroupStatus warning+unhealthy = %s; want bg-red-500", got)
	}
}

func TestFormatSignalDetails(t *testing.T) {
	sig := signal.Signal{Details: []signal.Detail{
		{Title: "Summary", Type: signal.DetailTypeText, Text: "hello"},
		{Title: "Object", Type: signal.DetailTypeObject, Object: map[string]any{"count": float64(3)}},
		{Title: "Table", Type: signal.DetailTypeTable, Table: &signal.TableDetail{Columns: []string{"name", "count"}, Rows: [][]string{{"jobs", "3"}}}},
		{Title: "List", Type: signal.DetailTypeList, List: []string{"a", "b"}},
	}}

	got := formatSignalDetails(sig)
	for _, want := range []string{
		"Summary: hello",
		"Object: {\"count\":3}",
		"Table: name | count\njobs | 3",
		"List: a, b",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("formatSignalDetails() = %q; missing %q", got, want)
		}
	}
}

func TestTreemap_RootAndPrefixGroupingCounts(t *testing.T) {
	fixtures := makeFixtures()

	root := buildTreemapData(fixtures, 0, "/")
	gotRootGroups := map[string]bool{}
	for _, g := range root.Groups {
		gotRootGroups[g.Name] = true
	}
	if !gotRootGroups["GroupA"] || !gotRootGroups["GroupB"] || len(gotRootGroups) != 2 {
		t.Fatalf("root groups = %#v; want exactly GroupA and GroupB", gotRootGroups)
	}
	if len(root.ThisLevel) != 0 {
		t.Fatalf("root ThisLevel size = %d; want 0", len(root.ThisLevel))
	}

	auto := buildTreemapData(fixtures, 1, "/GroupA")
	gotAutoGroups := map[string]bool{}
	for _, g := range auto.Groups {
		gotAutoGroups[g.Name] = true
	}
	if !gotAutoGroups["Env1"] || !gotAutoGroups["Env2"] || len(gotAutoGroups) != 2 {
		t.Fatalf("/GroupA groups = %#v; want exactly Env1 and Env2", gotAutoGroups)
	}
	if len(auto.ThisLevel) != 0 {
		t.Fatalf("/GroupA ThisLevel size = %d; want 0", len(auto.ThisLevel))
	}
}

func TestTreemap_Prefix_NoLeakAndCounts(t *testing.T) {
	fixtures := makeFixtures()

	lb := buildTreemapData(fixtures, 1, "/GroupB")

	if len(lb.Groups) != 0 {
		t.Fatalf("/GroupB groups = %d; want 0", len(lb.Groups))
	}
	if len(lb.ThisLevel) != 2 {
		t.Fatalf("/GroupB ThisLevel size = %d; want 2", len(lb.ThisLevel))
	}
	names := map[string]bool{}
	for _, a := range lb.ThisLevel {
		names[a.Alarm.Name] = true
	}
	if !names["Metric One"] || !names["Metric Two"] {
		t.Fatalf("/GroupB ThisLevel names = %#v; missing expected alarms", names)
	}
}
