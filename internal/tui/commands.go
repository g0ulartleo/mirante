package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/apiclient"
)

func marqueeTickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*200, func(t time.Time) tea.Msg {
		return marqueeTickMsg{}
	})
}

func fetchDetailCmd(client *apiclient.Client, id string) tea.Cmd {
	return func() tea.Msg {
		signals, err := client.GetAlarmSignals(id, detailSignalsFetchLimit)
		if err != nil {
			return errMsg{err}
		}
		return detailSignalsMsg{id: id, signals: signals}
	}
}

func fetchCmd(client *apiclient.Client) tea.Cmd {
	return func() tea.Msg {
		data, err := client.GetAllAlarmSignals()
		if err != nil {
			return errMsg{err}
		}
		return signalsMsg(data)
	}
}

func connectCmd(ctx context.Context, client *apiclient.Client) tea.Cmd {
	return func() tea.Msg {
		updates, errs, err := client.SubscribeAlarmSignals(ctx)
		if err != nil {
			return errMsg{err}
		}
		return connectedMsg{updates: updates, errs: errs}
	}
}

func listenCmd(ctx context.Context, updates <-chan []alarm.AlarmSignals, errs <-chan error) tea.Cmd {
	return func() tea.Msg {
		select {
		case <-ctx.Done():
			return wsClosedMsg{}
		case err := <-errs:
			if err != nil {
				return errMsg{err}
			}
			return wsClosedMsg{}
		case data, ok := <-updates:
			if !ok {
				return wsClosedMsg{}
			}
			return wsUpdateMsg(data)
		}
	}
}

func syncAlarmsCmd(client *apiclient.Client) tea.Cmd {
	return func() tea.Msg {
		if err := client.SyncAlarms(); err != nil {
			return errMsg{err}
		}
		return successMsg{"synced"}
	}
}

func runAlarmCmd(client *apiclient.Client, id string) tea.Cmd {
	return func() tea.Msg {
		if err := client.RunAlarm(id); err != nil {
			return errMsg{err}
		}
		return successMsg{"running…"}
	}
}

func reconnectCmd(ctx context.Context, client *apiclient.Client) tea.Cmd {
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		updates, errs, err := client.SubscribeAlarmSignals(ctx)
		if err != nil {
			return errMsg{err}
		}
		return connectedMsg{updates: updates, errs: errs}
	})
}
