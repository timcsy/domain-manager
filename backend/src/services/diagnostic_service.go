package services

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/domain-manager/backend/src/db"
	"github.com/domain-manager/backend/src/k8s"
	"github.com/domain-manager/backend/src/models"
)

// DiagnosticService handles system diagnostics and monitoring
type DiagnosticService struct{}

// NewDiagnosticService creates a new diagnostic service
func NewDiagnosticService() *DiagnosticService {
	return &DiagnosticService{}
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Status     string                 `json:"status"`
	Timestamp  time.Time              `json:"timestamp"`
	Components map[string]ComponentHealth `json:"components"`
	System     SystemInfo             `json:"system"`
}

// ComponentHealth represents the health of a component
type ComponentHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// SystemInfo contains system information
type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCPU       int    `json:"num_cpu"`
	Uptime       string `json:"uptime"`
}

var startTime = time.Now()

// GetHealthStatus performs comprehensive health check
func (s *DiagnosticService) GetHealthStatus() (*HealthCheckResult, error) {
	result := &HealthCheckResult{
		Timestamp:  time.Now(),
		Components: make(map[string]ComponentHealth),
		System: SystemInfo{
			GoVersion:    runtime.Version(),
			NumGoroutine: runtime.NumGoroutine(),
			NumCPU:       runtime.NumCPU(),
			Uptime:       time.Since(startTime).Round(time.Second).String(),
		},
	}

	// Check database
	if err := db.Health(); err != nil {
		result.Components["database"] = ComponentHealth{
			Status:  "unhealthy",
			Message: err.Error(),
		}
		result.Status = "degraded"
	} else {
		result.Components["database"] = ComponentHealth{
			Status: "healthy",
		}
	}

	// Check Kubernetes connection
	healthChecker := k8s.NewHealthChecker()
	if err := healthChecker.Check(); err != nil {
		result.Components["kubernetes"] = ComponentHealth{
			Status:  "unhealthy",
			Message: err.Error(),
		}
		if result.Status != "degraded" {
			result.Status = "degraded"
		}
	} else {
		result.Components["kubernetes"] = ComponentHealth{
			Status: "healthy",
		}
	}

	// Overall status
	if result.Status == "" {
		result.Status = "healthy"
	}

	return result, nil
}

// LogEntry represents a log entry
type LogEntry struct {
	ID        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Category  string    `json:"category"`
	Message   string    `json:"message"`
	Metadata  string    `json:"metadata,omitempty"`
}

// GetLogs retrieves system logs
func (s *DiagnosticService) GetLogs(filter models.LogFilter) ([]LogEntry, int, error) {
	query := `
		SELECT id, created_at as timestamp, log_type as level, category, message, details as metadata
		FROM diagnostic_logs
		WHERE 1=1
	`
	args := []interface{}{}

	// Apply filters
	if filter.Level != "" {
		query += " AND log_type = ?"
		args = append(args, filter.Level)
	}
	if filter.Category != "" {
		query += " AND category = ?"
		args = append(args, filter.Category)
	}
	if !filter.StartTime.IsZero() {
		query += " AND created_at >= ?"
		args = append(args, filter.StartTime)
	}
	if !filter.EndTime.IsZero() {
		query += " AND created_at <= ?"
		args = append(args, filter.EndTime)
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM (" + query + ")"
	var total int
	err := db.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		log.Printf("Failed to count logs: %v", err)
		return nil, 0, fmt.Errorf("failed to count logs: %w", err)
	}

	// Add ordering and pagination
	query += " ORDER BY created_at DESC"
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		log.Printf("Failed to query logs: %v", err)
		return nil, 0, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var entry LogEntry
		var metadata *string
		err := rows.Scan(
			&entry.ID,
			&entry.Timestamp,
			&entry.Level,
			&entry.Category,
			&entry.Message,
			&metadata,
		)
		if err != nil {
			log.Printf("Failed to scan log entry: %v", err)
			continue
		}
		if metadata != nil {
			entry.Metadata = *metadata
		}
		logs = append(logs, entry)
	}

	if logs == nil {
		logs = []LogEntry{}
	}

	return logs, total, nil
}

// GetSystemMetrics retrieves system metrics
type SystemMetrics struct {
	Domains      MetricCount `json:"domains"`
	Certificates MetricCount `json:"certificates"`
	Services     MetricCount `json:"services"`
	Memory       MemoryStats `json:"memory"`
}

type MetricCount struct {
	Total  int `json:"total"`
	Active int `json:"active"`
}

type MemoryStats struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"total_alloc"`
	Sys        uint64 `json:"sys"`
	NumGC      uint32 `json:"num_gc"`
}

func (s *DiagnosticService) GetSystemMetrics() (*SystemMetrics, error) {
	metrics := &SystemMetrics{}

	// Count domains
	err := db.DB.QueryRow("SELECT COUNT(*) FROM domains").Scan(&metrics.Domains.Total)
	if err != nil {
		log.Printf("Failed to count domains: %v", err)
	}
	err = db.DB.QueryRow("SELECT COUNT(*) FROM domains WHERE status = 'active'").Scan(&metrics.Domains.Active)
	if err != nil {
		log.Printf("Failed to count active domains: %v", err)
	}

	// Count certificates
	err = db.DB.QueryRow("SELECT COUNT(*) FROM certificates").Scan(&metrics.Certificates.Total)
	if err != nil {
		log.Printf("Failed to count certificates: %v", err)
	}
	err = db.DB.QueryRow("SELECT COUNT(*) FROM certificates WHERE status = 'valid'").Scan(&metrics.Certificates.Active)
	if err != nil {
		log.Printf("Failed to count valid certificates: %v", err)
	}

	// Get memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	metrics.Memory = MemoryStats{
		Alloc:      m.Alloc / 1024 / 1024,      // MB
		TotalAlloc: m.TotalAlloc / 1024 / 1024, // MB
		Sys:        m.Sys / 1024 / 1024,        // MB
		NumGC:      m.NumGC,
	}

	return metrics, nil
}
