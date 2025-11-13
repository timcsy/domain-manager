package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/domain-manager/backend/src/services"
)

// CertificateMonitorConfig holds configuration for certificate monitoring
type CertificateMonitorConfig struct {
	CheckInterval     time.Duration // How often to check certificates
	RenewalThreshold  int           // Days before expiry to trigger renewal
	Enabled           bool          // Enable/disable monitoring
}

// DefaultCertificateMonitorConfig returns default configuration
func DefaultCertificateMonitorConfig() *CertificateMonitorConfig {
	return &CertificateMonitorConfig{
		CheckInterval:    24 * time.Hour, // Check daily
		RenewalThreshold: 30,              // Renew 30 days before expiry
		Enabled:          true,
	}
}

// CertificateMonitor manages certificate monitoring and renewal
type CertificateMonitor struct {
	config  *CertificateMonitorConfig
	certSvc *services.CertificateService
}

// NewCertificateMonitor creates a new certificate monitor
func NewCertificateMonitor(
	config *CertificateMonitorConfig,
	certSvc *services.CertificateService,
) *CertificateMonitor {
	if config == nil {
		config = DefaultCertificateMonitorConfig()
	}

	return &CertificateMonitor{
		config:  config,
		certSvc: certSvc,
	}
}

// CheckAndRenewCertificates checks for expiring certificates and renews them
func (m *CertificateMonitor) CheckAndRenewCertificates(ctx context.Context) error {
	if !m.config.Enabled {
		log.Println("CertMonitor: Monitoring is disabled")
		return nil
	}

	log.Printf("CertMonitor: Checking certificates (threshold: %d days)", m.config.RenewalThreshold)

	// Use certificate service's auto-renewal function
	err := m.certSvc.AutoRenewCertificates(m.config.RenewalThreshold)
	if err != nil {
		log.Printf("CertMonitor: Auto-renewal failed: %v", err)
		return err
	}

	log.Println("CertMonitor: Certificate check completed")
	return nil
}

// RegisterWithScheduler registers the monitor as a scheduled task
func (m *CertificateMonitor) RegisterWithScheduler(scheduler *Scheduler) {
	if scheduler == nil {
		log.Println("CertMonitor: Cannot register with nil scheduler")
		return
	}

	if !m.config.Enabled {
		log.Println("CertMonitor: Not registering (monitoring disabled)")
		return
	}

	task := &Task{
		Name:     "CertificateRenewal",
		Interval: m.config.CheckInterval,
		Fn:       m.CheckAndRenewCertificates,
	}

	scheduler.AddTask(task)
	log.Printf("CertMonitor: Registered with scheduler (interval: %v)", m.config.CheckInterval)
}

// Enable enables certificate monitoring
func (m *CertificateMonitor) Enable() {
	m.config.Enabled = true
	log.Println("CertMonitor: Monitoring enabled")
}

// Disable disables certificate monitoring
func (m *CertificateMonitor) Disable() {
	m.config.Enabled = false
	log.Println("CertMonitor: Monitoring disabled")
}

// IsEnabled returns whether monitoring is enabled
func (m *CertificateMonitor) IsEnabled() bool {
	return m.config.Enabled
}

// SetCheckInterval sets the check interval
func (m *CertificateMonitor) SetCheckInterval(interval time.Duration) {
	m.config.CheckInterval = interval
	log.Printf("CertMonitor: Check interval set to %v", interval)
}

// SetRenewalThreshold sets the renewal threshold in days
func (m *CertificateMonitor) SetRenewalThreshold(days int) {
	m.config.RenewalThreshold = days
	log.Printf("CertMonitor: Renewal threshold set to %d days", days)
}

// GetConfig returns the current configuration
func (m *CertificateMonitor) GetConfig() *CertificateMonitorConfig {
	return m.config
}
