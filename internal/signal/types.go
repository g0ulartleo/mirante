package signal

import "time"

type Signal struct {
	AlarmID   string
	Status    Status
	Timestamp time.Time
	Message   string
	Details   []Detail
}

type Detail struct {
	Title  string
	Type   DetailType
	Text   string
	Object map[string]any
	Table  *TableDetail
	List   []string
}

type DetailType string

const (
	DetailTypeText   DetailType = "text"
	DetailTypeObject DetailType = "object"
	DetailTypeTable  DetailType = "table"
	DetailTypeList   DetailType = "list"
)

type TableDetail struct {
	Columns []string
	Rows    [][]string
}

type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusUnknown   Status = "unknown"
	StatusWarning   Status = "warning"
)
