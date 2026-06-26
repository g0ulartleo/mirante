package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/apiclient"
	"github.com/g0ulartleo/mirante/internal/signal"
)

type viewMode int

const (
	listView viewMode = iota
	detailView
)

type listMode int

const (
	triageList listMode = iota
	allList
)

func (l listMode) String() string {
	switch l {
	case allList:
		return "all"
	default:
		return "triage"
	}
}

type inputMode int

const (
	inputNone inputMode = iota
	inputFilter
	inputCommand
)

const detailSignalsPageSize = 10
const detailSignalsFetchLimit = 100

type Model struct {
	client *apiclient.Client

	ctx     context.Context
	cancel  context.CancelFunc
	updates <-chan []alarm.AlarmSignals
	errs    <-chan error

	all  []alarm.AlarmSignals
	mode viewMode

	rows        []listRow
	alarmRows   []int
	alarmCursor int
	topRow      int
	filter      string
	sort        sortMode
	listMode    listMode

	input    inputMode
	inputBuf string

	selectedID      string
	detail          viewport.Model
	detailReady     bool
	detailSignals   []signal.Signal
	detailHistoryAt time.Time
	detailLoading   bool

	width      int
	height     int
	lastUpdate time.Time
	connected  bool
	err        error
	statusMsg  string
	showHelp   bool
	leaderMode bool

	commandPaletteOpen   bool
	commandPaletteCursor int

	marqueeOffset int

	detailSignalCursor int
}

func NewModel(client *apiclient.Client) *Model {
	ctx, cancel := context.WithCancel(context.Background())
	return &Model{
		client:   client,
		ctx:      ctx,
		cancel:   cancel,
		mode:     listView,
		sort:     sortSeverity,
		listMode: triageList,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(fetchCmd(m.client), connectCmd(m.ctx, m.client), marqueeTickCmd())
}
