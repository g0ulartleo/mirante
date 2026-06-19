package tui

func (m *Model) View() string {
	if m.width == 0 {
		return "Loading…"
	}
	var s string
	if m.mode == detailView {
		s = m.detailScreen()
	} else {
		s = m.listScreen()
	}
	if m.showHelp {
		s = m.overlayHelp()
	}
	return s
}
