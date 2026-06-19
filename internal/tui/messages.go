package tui

import (
	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/signal"
)

type marqueeTickMsg struct{}

type detailSignalsMsg struct {
	id      string
	signals []signal.Signal
}

type signalsMsg []alarm.AlarmSignals

type wsUpdateMsg []alarm.AlarmSignals

type connectedMsg struct {
	updates <-chan []alarm.AlarmSignals
	errs    <-chan error
}

type errMsg struct{ err error }

type successMsg struct{ text string }

type wsClosedMsg struct{}
