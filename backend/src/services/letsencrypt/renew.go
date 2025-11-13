package letsencrypt

import (
	"fmt"
	"log"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
)

// RenewalOptions contains options for renewing a certificate
type RenewalOptions struct {
	CertPEM    []byte        // Current certificate PEM
	PrivateKey []byte        // Current private key PEM
	Domains    []string      // Domains to renew (should match original)
	HTTPPort   int           // Port for HTTP-01 challenge
	Timeout    time.Duration // Challenge timeout
}

// RenewalResult contains information about the renewal
type RenewalResult struct {
	Renewed        bool      // Whether renewal was performed
	NewCertificate []byte    // New certificate PEM (if renewed)
	NewPrivateKey  []byte    // New private key PEM (if renewed)
	ExpiresAt      time.Time // New expiration date
	Reason         string    // Reason for renewal or skip
}

const (
	// DefaultRenewalWindow is the default time before expiry to renew (30 days)
	DefaultRenewalWindow = 30 * 24 * time.Hour

	// MinRenewalWindow is the minimum time before expiry (7 days)
	MinRenewalWindow = 7 * 24 * time.Hour
)

// RenewCertificate renews a certificate if it's close to expiry
func (c *Client) RenewCertificate(opts *RenewalOptions) (*RenewalResult, error) {
	if opts.CertPEM == nil || len(opts.CertPEM) == 0 {
		return nil, fmt.Errorf("certificate PEM is required")
	}

	// Parse current certificate to check expiry
	cert, err := certcrypto.ParsePEMCertificate(opts.CertPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Check if renewal is needed
	needsRenewal, daysLeft := NeedsRenewal(cert.NotAfter, DefaultRenewalWindow)
	if !needsRenewal {
		log.Printf("Certificate for %v doesn't need renewal yet (%d days left)", opts.Domains, daysLeft)
		return &RenewalResult{
			Renewed:   false,
			ExpiresAt: cert.NotAfter,
			Reason:    fmt.Sprintf("Certificate valid for %d more days", daysLeft),
		}, nil
	}

	log.Printf("Renewing certificate for %v (%d days left)", opts.Domains, daysLeft)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Create certificate resource for renewal
	certResource := &certificate.Resource{
		Domain:            opts.Domains[0],
		Certificate:       opts.CertPEM,
		PrivateKey:        opts.PrivateKey,
		IssuerCertificate: nil, // Not required for renewal
	}

	// Perform renewal using the new RenewWithOptions method
	newCert, err := c.client.Certificate.RenewWithOptions(*certResource, &certificate.RenewOptions{
		Bundle:                  true,
		PreferredChain:          "",
		AlwaysDeactivateAuthorizations: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to renew certificate: %w", err)
	}

	// Parse new certificate to get expiry
	newCertParsed, err := certcrypto.ParsePEMCertificate(newCert.Certificate)
	if err != nil {
		log.Printf("Warning: failed to parse renewed certificate: %v", err)
	}

	result := &RenewalResult{
		Renewed:        true,
		NewCertificate: newCert.Certificate,
		NewPrivateKey:  newCert.PrivateKey,
		ExpiresAt:      newCertParsed.NotAfter,
		Reason:         fmt.Sprintf("Renewed (was expiring in %d days)", daysLeft),
	}

	log.Printf("Successfully renewed certificate for %v (new expiry: %s)",
		opts.Domains, result.ExpiresAt.Format("2006-01-02"))

	return result, nil
}

// RenewCertificateForce forces renewal regardless of expiry date
func (c *Client) RenewCertificateForce(opts *RenewalOptions) (*RenewalResult, error) {
	log.Printf("Forcing certificate renewal for %v", opts.Domains)

	// Simply obtain a new certificate
	obtainOpts := &ObtainOptions{
		Domains:  opts.Domains,
		HTTPPort: opts.HTTPPort,
		Timeout:  opts.Timeout,
	}

	newCert, err := c.ObtainCertificate(obtainOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain new certificate: %w", err)
	}

	return &RenewalResult{
		Renewed:        true,
		NewCertificate: newCert.Certificate,
		NewPrivateKey:  newCert.PrivateKey,
		ExpiresAt:      newCert.NotAfter,
		Reason:         "Forced renewal",
	}, nil
}

// NeedsRenewal checks if a certificate needs renewal based on expiry date
func NeedsRenewal(expiresAt time.Time, renewalWindow time.Duration) (bool, int) {
	now := time.Now()
	timeLeft := expiresAt.Sub(now)
	daysLeft := int(timeLeft.Hours() / 24)

	if timeLeft < 0 {
		// Already expired
		return true, daysLeft
	}

	if renewalWindow == 0 {
		renewalWindow = DefaultRenewalWindow
	}

	// Renew if within renewal window
	needsRenewal := timeLeft <= renewalWindow

	return needsRenewal, daysLeft
}

// CalculateRenewalDate calculates when a certificate should be renewed
func CalculateRenewalDate(expiresAt time.Time, renewalWindow time.Duration) time.Time {
	if renewalWindow == 0 {
		renewalWindow = DefaultRenewalWindow
	}

	return expiresAt.Add(-renewalWindow)
}

// IsExpired checks if a certificate has expired
func IsExpired(expiresAt time.Time) bool {
	return time.Now().After(expiresAt)
}

// IsExpiringSoon checks if a certificate is expiring within the given duration
func IsExpiringSoon(expiresAt time.Time, within time.Duration) bool {
	if within == 0 {
		within = DefaultRenewalWindow
	}

	return time.Until(expiresAt) <= within
}

// GetExpiryInfo returns human-readable expiry information
func GetExpiryInfo(expiresAt time.Time) string {
	now := time.Now()

	if now.After(expiresAt) {
		elapsed := now.Sub(expiresAt)
		days := int(elapsed.Hours() / 24)
		return fmt.Sprintf("Expired %d days ago", days)
	}

	timeLeft := expiresAt.Sub(now)
	days := int(timeLeft.Hours() / 24)

	if days == 0 {
		hours := int(timeLeft.Hours())
		return fmt.Sprintf("Expires in %d hours", hours)
	}

	return fmt.Sprintf("Expires in %d days", days)
}

// BatchRenewalStatus tracks renewal status for multiple certificates
type BatchRenewalStatus struct {
	TotalCerts     int
	RenewedCerts   int
	SkippedCerts   int
	FailedCerts    int
	Failures       []RenewalFailure
}

// RenewalFailure tracks a failed renewal attempt
type RenewalFailure struct {
	Domain string
	Error  string
}

// ShouldRenewBatch determines if any certificates in a batch need renewal
func ShouldRenewBatch(certificates []time.Time, window time.Duration) bool {
	for _, expiresAt := range certificates {
		if needsRenewal, _ := NeedsRenewal(expiresAt, window); needsRenewal {
			return true
		}
	}
	return false
}
