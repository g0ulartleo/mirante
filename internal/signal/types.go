package signal

import "time"

type Signal struct {
	AlarmID   string
	Status    Status
	Timestamp time.Time
	Message   string
	Details   []map[string]any
}

type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusUnknown   Status = "unknown"
	StatusWarning   Status = "warning"
)
