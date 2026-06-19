package tui

import (
	"fmt"
	"time"
)

func humanizeDuration(d time.Duration) string {
	seconds := int(d.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	if hours < 24 {
		return fmt.Sprintf("%dh", hours)
	}
	days := hours / 24
	return fmt.Sprintf("%dd", days)
}

func ageString(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return humanizeDuration(time.Since(t)) + " ago"
}

func ageShort(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return humanizeDuration(time.Since(t))
}
