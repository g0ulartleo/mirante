package tui

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/g0ulartleo/mirante/internal/signal"
)

func copySignalCmd(s signal.Signal) tea.Cmd {
	return func() tea.Msg {
		text := signalDetailText(s)
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err != nil {
			return errMsg{err}
		}
		return successMsg{text: "copied alarm signal ✓"}
	}
}

func signalDetailText(s signal.Signal) string {
	text := fmt.Sprintf("Status: %s\nTime: %s\nMessage: %s\n", s.Status, s.Timestamp.Format(time.RFC3339), s.Message)
	if len(s.Details) > 0 {
		data, err := json.MarshalIndent(s.Details, "", "  ")
		if err == nil {
			text += "Details:\n" + string(data) + "\n"
		}
	}
	return text
}

func copyAlarmIDCmd(id string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(id)
		if err := cmd.Run(); err != nil {
			return errMsg{err}
		}
		return successMsg{text: "copied alarm id ✓"}
	}
}
