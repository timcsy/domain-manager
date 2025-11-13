package letsencrypt

import (
	"fmt"
	"log"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/http01"
)

// CertificateResult contains the obtained certificate and private key
type CertificateResult struct {
	Certificate []byte    // PEM-encoded certificate chain
	PrivateKey  []byte    // PEM-encoded private key
	IssuerCert  []byte    // PEM-encoded issuer certificate
	Domain      string    // Primary domain
	NotBefore   time.Time // Certificate validity start
	NotAfter    time.Time // Certificate validity end
}

// ObtainOptions contains options for obtaining a certificate
type ObtainOptions struct {
	Domains         []string      // List of domains to obtain certificate for
	HTTPPort        int           // Port for HTTP-01 challenge (default: 80)
	Timeout         time.Duration // Challenge timeout (default: 30s)
	MustStaple      bool          // Request OCSP Must-Staple
}

// ObtainCertificate obtains a new certificate from Let's Encrypt
func (c *Client) ObtainCertificate(opts *ObtainOptions) (*CertificateResult, error) {
	if len(opts.Domains) == 0 {
		return nil, fmt.Errorf("at least one domain is required")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Set default options
	if opts.HTTPPort == 0 {
		opts.HTTPPort = 80
	}
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}

	log.Printf("Obtaining certificate for domains: %v (staging: %v)", opts.Domains, c.staging)

	// Configure HTTP-01 challenge
	httpProvider := http01.NewProviderServer("", fmt.Sprintf("%d", opts.HTTPPort))
	if err := c.client.Challenge.SetHTTP01Provider(httpProvider); err != nil {
		return nil, fmt.Errorf("failed to set HTTP-01 provider: %w", err)
	}

	// Prepare certificate request
	request := certificate.ObtainRequest{
		Domains: opts.Domains,
		Bundle:  true, // Include full certificate chain
	}

	// Set OCSP Must-Staple if requested
	if opts.MustStaple {
		request.MustStaple = true
	}

	// Obtain certificate
	cert, err := c.client.Certificate.Obtain(request)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain certificate: %w", err)
	}

	// Parse certificate to get validity dates
	notBefore, notAfter, err := parseCertificateDates(cert.Certificate)
	if err != nil {
		log.Printf("Warning: failed to parse certificate dates: %v", err)
		notBefore = time.Now()
		notAfter = time.Now().Add(90 * 24 * time.Hour) // Let's Encrypt default: 90 days
	}

	result := &CertificateResult{
		Certificate: cert.Certificate,
		PrivateKey:  cert.PrivateKey,
		IssuerCert:  cert.IssuerCertificate,
		Domain:      cert.Domain,
		NotBefore:   notBefore,
		NotAfter:    notAfter,
	}

	log.Printf("Successfully obtained certificate for %s (valid until: %s)",
		result.Domain, result.NotAfter.Format("2006-01-02"))

	return result, nil
}

// ObtainCertificateWithRetry obtains a certificate with automatic retry
func (c *Client) ObtainCertificateWithRetry(opts *ObtainOptions, maxRetries int) (*CertificateResult, error) {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			backoff := time.Duration(i*i) * time.Second // Exponential backoff
			log.Printf("Retrying certificate request in %v (attempt %d/%d)", backoff, i+1, maxRetries)
			time.Sleep(backoff)
		}

		result, err := c.ObtainCertificate(opts)
		if err == nil {
			return result, nil
		}

		lastErr = err
		log.Printf("Certificate request failed (attempt %d/%d): %v", i+1, maxRetries, err)
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// ObtainWildcardCertificate obtains a wildcard certificate (requires DNS-01 challenge)
// Note: This is a placeholder - DNS-01 challenge requires DNS provider integration
func (c *Client) ObtainWildcardCertificate(domain string) (*CertificateResult, error) {
	// Wildcard certificates require DNS-01 challenge
	// This would need DNS provider integration (e.g., Route53, CloudFlare, etc.)
	return nil, fmt.Errorf("wildcard certificates require DNS-01 challenge (not yet implemented)")
}

// parseCertificateDates extracts validity dates from PEM-encoded certificate
func parseCertificateDates(certPEM []byte) (notBefore, notAfter time.Time, err error) {
	cert, err := certcrypto.ParsePEMCertificate(certPEM)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return cert.NotBefore, cert.NotAfter, nil
}

// ValidateDomain performs basic domain validation checks
func ValidateDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	// Check for wildcard (requires DNS-01)
	if len(domain) > 0 && domain[0] == '*' {
		return fmt.Errorf("wildcard domains require DNS-01 challenge (not yet implemented)")
	}

	// Basic domain format validation
	if len(domain) > 253 {
		return fmt.Errorf("domain name too long (max 253 characters)")
	}

	return nil
}

// GetChallengePort returns the HTTP port used for HTTP-01 challenge
func GetChallengePort() int {
	return 80 // HTTP-01 challenge always uses port 80
}
