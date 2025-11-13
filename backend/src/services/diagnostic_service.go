package services

import (
	"fmt"
	"log"
	"net/http"
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

// DomainDiagnostics represents diagnostic information for a domain
type DomainDiagnostics struct {
	Domain          string              `json:"domain"`
	Status          string              `json:"status"`
	Ingress         IngressDiagnostic   `json:"ingress"`
	Service         ServiceDiagnostic   `json:"service"`
	Certificate     CertificateDiagnostic `json:"certificate"`
	OverallHealth   string              `json:"overall_health"`
	LastChecked     time.Time           `json:"last_checked"`
}

// IngressDiagnostic represents ingress diagnostic information
type IngressDiagnostic struct {
	Status      string `json:"status"`      // "healthy", "warning", "error", "not_found"
	Message     string `json:"message"`
	Configured  bool   `json:"configured"`
	Host        string `json:"host,omitempty"`
}

// ServiceDiagnostic represents service diagnostic information
type ServiceDiagnostic struct {
	Status      string `json:"status"`      // "healthy", "warning", "error", "not_found"
	Message     string `json:"message"`
	Reachable   bool   `json:"reachable"`
	Namespace   string `json:"namespace,omitempty"`
	Name        string `json:"name,omitempty"`
	Port        int    `json:"port,omitempty"`
}

// CertificateDiagnostic represents certificate diagnostic information
type CertificateDiagnostic struct {
	Status      string     `json:"status"`      // "healthy", "warning", "error", "not_configured"
	Message     string     `json:"message"`
	Configured  bool       `json:"configured"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	DaysRemaining int      `json:"days_remaining,omitempty"`
}

// GetDomainDiagnostics performs comprehensive diagnostics for a domain
func (s *DiagnosticService) GetDomainDiagnostics(domainID int64) (*DomainDiagnostics, error) {
	// Get domain from database
	var domain models.Domain
	err := db.DB.QueryRow(`
		SELECT id, domain_name, target_service, target_namespace, target_port,
		       ssl_mode, certificate_id, status, enabled
		FROM domains
		WHERE id = ?
	`, domainID).Scan(
		&domain.ID,
		&domain.DomainName,
		&domain.TargetService,
		&domain.TargetNamespace,
		&domain.TargetPort,
		&domain.SSLMode,
		&domain.CertificateID,
		&domain.Status,
		&domain.Enabled,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}

	diagnostics := &DomainDiagnostics{
		Domain:      domain.DomainName,
		Status:      domain.Status,
		LastChecked: time.Now(),
	}

	// Check Ingress status
	diagnostics.Ingress = s.checkIngressStatus(&domain)

	// Check Service status
	diagnostics.Service = s.checkServiceStatus(&domain)

	// Check Certificate status
	diagnostics.Certificate = s.checkCertificateStatus(&domain)

	// Determine overall health
	diagnostics.OverallHealth = s.determineOverallHealth(diagnostics)

	return diagnostics, nil
}

// checkIngressStatus checks the ingress configuration status
func (s *DiagnosticService) checkIngressStatus(domain *models.Domain) IngressDiagnostic {
	ingressMgr := k8s.NewIngressManager()
	ingressName := fmt.Sprintf("domain-%d", domain.ID)

	ingress, err := ingressMgr.GetIngress(domain.TargetNamespace, ingressName)
	if err != nil {
		return IngressDiagnostic{
			Status:     "not_found",
			Message:    fmt.Sprintf("Ingress not found: %v", err),
			Configured: false,
		}
	}

	// Check if ingress has load balancer IP
	if len(ingress.Status.LoadBalancer.Ingress) == 0 {
		return IngressDiagnostic{
			Status:     "warning",
			Message:    "Ingress exists but no load balancer IP assigned yet",
			Configured: true,
			Host:       domain.DomainName,
		}
	}

	return IngressDiagnostic{
		Status:     "healthy",
		Message:    "Ingress is configured and active",
		Configured: true,
		Host:       domain.DomainName,
	}
}

// checkServiceStatus checks the target service status
func (s *DiagnosticService) checkServiceStatus(domain *models.Domain) ServiceDiagnostic {
	serviceMgr := k8s.NewServiceManager()

	// Check if service exists
	exists, err := serviceMgr.ServiceExists(domain.TargetNamespace, domain.TargetService)
	if err != nil {
		return ServiceDiagnostic{
			Status:    "error",
			Message:   fmt.Sprintf("Failed to check service: %v", err),
			Reachable: false,
		}
	}

	if !exists {
		return ServiceDiagnostic{
			Status:    "not_found",
			Message:   fmt.Sprintf("Service %s not found in namespace %s", domain.TargetService, domain.TargetNamespace),
			Reachable: false,
			Namespace: domain.TargetNamespace,
			Name:      domain.TargetService,
			Port:      domain.TargetPort,
		}
	}

	// Validate service port
	err = serviceMgr.ValidateService(domain.TargetNamespace, domain.TargetService, domain.TargetPort)
	if err != nil {
		return ServiceDiagnostic{
			Status:    "warning",
			Message:   fmt.Sprintf("Service exists but port validation failed: %v", err),
			Reachable: false,
			Namespace: domain.TargetNamespace,
			Name:      domain.TargetService,
			Port:      domain.TargetPort,
		}
	}

	return ServiceDiagnostic{
		Status:    "healthy",
		Message:   "Service is reachable and port is valid",
		Reachable: true,
		Namespace: domain.TargetNamespace,
		Name:      domain.TargetService,
		Port:      domain.TargetPort,
	}
}

// checkCertificateStatus checks the SSL certificate status
func (s *DiagnosticService) checkCertificateStatus(domain *models.Domain) CertificateDiagnostic {
	if domain.CertificateID == nil {
		if domain.SSLMode == "auto" {
			// For auto mode, check if cert-manager has created the TLS secret
			secretMgr := k8s.NewSecretManager()
			secretName := fmt.Sprintf("domain-%d-tls", domain.ID)
			exists, err := secretMgr.SecretExists(domain.TargetNamespace, secretName)

			if err != nil {
				return CertificateDiagnostic{
					Status:     "warning",
					Message:    fmt.Sprintf("Unable to check certificate status: %v", err),
					Configured: false,
				}
			}

			if exists {
				// Get the secret to check certificate expiry
				secret, err := secretMgr.GetSecret(domain.TargetNamespace, secretName)
				if err == nil && secret.Type == "kubernetes.io/tls" {
					// For cert-manager managed certificates, we can't easily get expiry without parsing
					// the certificate. For now, just report it as healthy.
					return CertificateDiagnostic{
						Status:     "healthy",
						Message:    "Certificate managed by cert-manager (auto SSL)",
						Configured: true,
					}
				}
			}

			return CertificateDiagnostic{
				Status:     "warning",
				Message:    "SSL mode is auto but certificate not yet issued by cert-manager",
				Configured: false,
			}
		}
		return CertificateDiagnostic{
			Status:     "not_configured",
			Message:    "No SSL certificate configured",
			Configured: false,
		}
	}

	// Get certificate from database
	var expiresAt time.Time
	var status string
	err := db.DB.QueryRow(`
		SELECT expires_at, status
		FROM certificates
		WHERE id = ?
	`, *domain.CertificateID).Scan(&expiresAt, &status)

	if err != nil {
		return CertificateDiagnostic{
			Status:     "error",
			Message:    fmt.Sprintf("Failed to get certificate: %v", err),
			Configured: true,
		}
	}

	daysRemaining := int(time.Until(expiresAt).Hours() / 24)

	if daysRemaining < 0 {
		return CertificateDiagnostic{
			Status:        "error",
			Message:       "Certificate has expired",
			Configured:    true,
			ExpiresAt:     &expiresAt,
			DaysRemaining: daysRemaining,
		}
	}

	if daysRemaining < 30 {
		return CertificateDiagnostic{
			Status:        "warning",
			Message:       fmt.Sprintf("Certificate expires in %d days", daysRemaining),
			Configured:    true,
			ExpiresAt:     &expiresAt,
			DaysRemaining: daysRemaining,
		}
	}

	return CertificateDiagnostic{
		Status:        "healthy",
		Message:       fmt.Sprintf("Certificate is valid (%d days remaining)", daysRemaining),
		Configured:    true,
		ExpiresAt:     &expiresAt,
		DaysRemaining: daysRemaining,
	}
}

// determineOverallHealth determines the overall health status
func (s *DiagnosticService) determineOverallHealth(diag *DomainDiagnostics) string {
	// If any component is in error state, overall health is error
	if diag.Ingress.Status == "error" ||
	   diag.Service.Status == "error" ||
	   diag.Certificate.Status == "error" {
		return "error"
	}

	// If any component is not found, it's a critical issue
	if diag.Ingress.Status == "not_found" ||
	   diag.Service.Status == "not_found" {
		return "error"
	}

	// If any component has warnings, overall health is degraded
	if diag.Ingress.Status == "warning" ||
	   diag.Service.Status == "warning" ||
	   diag.Certificate.Status == "warning" {
		return "degraded"
	}

	// If certificate is not configured, it's still operational but degraded
	if diag.Certificate.Status == "not_configured" {
		return "degraded"
	}

	return "healthy"
}

// CheckSubdomainHealth performs HTTP health check for a single subdomain
func (s *DiagnosticService) CheckSubdomainHealth(domainID int64, domainName string) (*models.HealthCheckResult, error) {
	startTime := time.Now()
	result := &models.HealthCheckResult{
		DomainID:   domainID,
		DomainName: domainName,
		CheckedAt:  startTime,
	}

	// Perform HTTP GET request to check subdomain health
	url := fmt.Sprintf("https://%s", domainName)
	client := &http.Client{
		Timeout: 10 * time.Second,
		// Don't follow redirects for health check
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		result.Healthy = false
		result.ErrorMessage = err.Error()
		result.ResponseTime = time.Since(startTime).Milliseconds()
		return result, nil
	}
	defer resp.Body.Close()

	result.ResponseTime = time.Since(startTime).Milliseconds()
	result.HTTPStatus = resp.StatusCode

	// Consider 2xx and 3xx as healthy
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		result.Healthy = true
	} else {
		result.Healthy = false
		result.ErrorMessage = fmt.Sprintf("HTTP status %d", resp.StatusCode)
	}

	return result, nil
}

// CheckAllSubdomains performs health check for all active subdomains
func (s *DiagnosticService) CheckAllSubdomains() (*models.SubdomainHealthSummary, error) {
	// Get all active domains from database
	rows, err := db.DB.Query(`
		SELECT id, domain_name
		FROM domains
		WHERE status = 'active' AND enabled = 1
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query domains: %w", err)
	}
	defer rows.Close()

	var domains []struct {
		ID   int64
		Name string
	}

	for rows.Next() {
		var d struct {
			ID   int64
			Name string
		}
		if err := rows.Scan(&d.ID, &d.Name); err != nil {
			log.Printf("Failed to scan domain: %v", err)
			continue
		}
		domains = append(domains, d)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating domains: %w", err)
	}

	// Perform health checks for all domains
	summary := &models.SubdomainHealthSummary{
		TotalSubdomains: len(domains),
		Results:         make([]models.HealthCheckResult, 0, len(domains)),
	}

	for _, domain := range domains {
		result, err := s.CheckSubdomainHealth(domain.ID, domain.Name)
		if err != nil {
			log.Printf("Failed to check health for %s: %v", domain.Name, err)
			continue
		}

		summary.Results = append(summary.Results, *result)
		if result.Healthy {
			summary.HealthySubdomains++
		} else {
			summary.UnhealthySubdomains++
		}
	}

	return summary, nil
}
