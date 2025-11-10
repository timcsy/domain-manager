package models

import "time"

// LogFilter represents filters for querying logs
type LogFilter struct {
	Level     string    // error, warn, info, debug
	Category  string    // domain, certificate, k8s, system, auth
	StartTime time.Time
	EndTime   time.Time
	Limit     int
	Offset    int
}
