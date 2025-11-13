package models

import (
	"time"
)

// DiagnosticLog represents a diagnostic log entry
type DiagnosticLog struct {
	ID          int64     `json:"id" db:"id"`
	DomainID    int64     `json:"domain_id" db:"domain_id"`
	DomainName  string    `json:"domain_name" db:"domain_name"`
	CheckType   string    `json:"check_type" db:"check_type"`     // "health", "ssl", "dns", "connectivity"
	Status      string    `json:"status" db:"status"`             // "success", "warning", "error"
	Message     string    `json:"message" db:"message"`
	Details     string    `json:"details,omitempty" db:"details"` // JSON string with additional details
	Resolved    bool      `json:"resolved" db:"resolved"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// DiagnosticLogFilter represents filters for listing diagnostic logs
type DiagnosticLogFilter struct {
	DomainID   *int64
	CheckType  string
	Status     string
	Resolved   *bool
	Limit      int
	Offset     int
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	DomainID      int64     `json:"domain_id"`
	DomainName    string    `json:"domain_name"`
	Healthy       bool      `json:"healthy"`
	HTTPStatus    int       `json:"http_status,omitempty"`
	ResponseTime  int64     `json:"response_time_ms,omitempty"` // milliseconds
	ErrorMessage  string    `json:"error_message,omitempty"`
	CheckedAt     time.Time `json:"checked_at"`
}

// SubdomainHealthSummary represents health summary for subdomains
type SubdomainHealthSummary struct {
	TotalSubdomains   int                  `json:"total_subdomains"`
	HealthySubdomains int                  `json:"healthy_subdomains"`
	UnhealthySubdomains int                `json:"unhealthy_subdomains"`
	Results           []HealthCheckResult  `json:"results"`
}
