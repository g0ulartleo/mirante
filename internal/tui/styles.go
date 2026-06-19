package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorBg     = lipgloss.Color("#0d1017")
	colorText   = lipgloss.Color("#c9d1d9")
	colorBright = lipgloss.Color("#e6edf3")
	colorMuted  = lipgloss.Color("#6e7681")
	colorFaint  = lipgloss.Color("#3b424d")
	colorBorder = lipgloss.Color("#262c36")

	colorHealthy   = lipgloss.Color("#7ee787")
	colorWarning   = lipgloss.Color("#e3b341")
	colorUnhealthy = lipgloss.Color("#ff7b9c")
	colorUnknown   = lipgloss.Color("#7d8590")
	colorAccent    = lipgloss.Color("#539bf5")
	colorSelBg     = lipgloss.Color("#243049")
)

var (
	panelTitleStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Bold(true)

	panelBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().Foreground(colorMuted)
	valueStyle = lipgloss.NewStyle().Foreground(colorText)

	brandStyle = lipgloss.NewStyle().Foreground(colorBright).Bold(true)

	colHeaderStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Bold(true)

	groupHeaderStyle = lipgloss.NewStyle().
				Foreground(colorBright).
				Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorUnhealthy).
			Bold(true).
			Padding(1, 2)

	breadcrumbStyle = lipgloss.NewStyle().Foreground(colorMuted)

	crumbActiveStyle = lipgloss.NewStyle().Foreground(colorBright)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	detailTitleStyle = lipgloss.NewStyle().Foreground(colorBright).Bold(true)
)
