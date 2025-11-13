package certificate

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"
)

// ValidationError represents a certificate validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationResult contains the results of certificate validation
type ValidationResult struct {
	Valid       bool
	Errors      []ValidationError
	Warnings    []string
	Certificate *x509.Certificate
	PrivateKey  crypto.PrivateKey
}

// CertificateValidator provides certificate validation functionality
type CertificateValidator struct{}

// NewValidator creates a new certificate validator
func NewValidator() *CertificateValidator {
	return &CertificateValidator{}
}

// ValidateCertificatePEM validates a certificate PEM format
func (v *CertificateValidator) ValidateCertificatePEM(certPEM string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []string{},
	}

	// Trim whitespace
	certPEM = strings.TrimSpace(certPEM)

	// Check if PEM is empty
	if certPEM == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "certificate_pem",
			Message: "certificate PEM cannot be empty",
		})
		return result, nil
	}

	// Decode PEM block
	block, rest := pem.Decode([]byte(certPEM))
	if block == nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "certificate_pem",
			Message: "failed to decode PEM block - not valid PEM format",
		})
		return result, nil
	}

	// Check PEM type
	if block.Type != "CERTIFICATE" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "certificate_pem",
			Message: fmt.Sprintf("invalid PEM type: expected 'CERTIFICATE', got '%s'", block.Type),
		})
		return result, nil
	}

	// Warn if there's extra data after the certificate
	if len(rest) > 0 && len(strings.TrimSpace(string(rest))) > 0 {
		result.Warnings = append(result.Warnings, "extra data found after certificate PEM block")
	}

	// Parse certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "certificate_pem",
			Message: fmt.Sprintf("failed to parse certificate: %v", err),
		})
		return result, nil
	}

	result.Certificate = cert

	// Validate certificate properties
	v.validateCertificateProperties(cert, result)

	return result, nil
}

// ValidatePrivateKeyPEM validates a private key PEM format
func (v *CertificateValidator) ValidatePrivateKeyPEM(keyPEM string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []string{},
	}

	// Trim whitespace
	keyPEM = strings.TrimSpace(keyPEM)

	// Check if PEM is empty
	if keyPEM == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "private_key_pem",
			Message: "private key PEM cannot be empty",
		})
		return result, nil
	}

	// Decode PEM block
	block, rest := pem.Decode([]byte(keyPEM))
	if block == nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "private_key_pem",
			Message: "failed to decode PEM block - not valid PEM format",
		})
		return result, nil
	}

	// Check PEM type
	validTypes := []string{
		"RSA PRIVATE KEY",
		"PRIVATE KEY",
		"EC PRIVATE KEY",
	}
	validType := false
	for _, t := range validTypes {
		if block.Type == t {
			validType = true
			break
		}
	}
	if !validType {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "private_key_pem",
			Message: fmt.Sprintf("invalid PEM type: expected one of %v, got '%s'", validTypes, block.Type),
		})
		return result, nil
	}

	// Warn if there's extra data after the key
	if len(rest) > 0 && len(strings.TrimSpace(string(rest))) > 0 {
		result.Warnings = append(result.Warnings, "extra data found after private key PEM block")
	}

	// Parse private key
	var privateKey crypto.PrivateKey
	var err error

	// Try PKCS1 (RSA)
	if block.Type == "RSA PRIVATE KEY" {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	} else if block.Type == "EC PRIVATE KEY" {
		// Try EC private key
		privateKey, err = x509.ParseECPrivateKey(block.Bytes)
	} else {
		// Try PKCS8 (generic)
		privateKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	}

	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "private_key_pem",
			Message: fmt.Sprintf("failed to parse private key: %v", err),
		})
		return result, nil
	}

	result.PrivateKey = privateKey

	// Validate key properties
	v.validatePrivateKeyProperties(privateKey, result)

	return result, nil
}

// ValidateCertificateKeyPair validates that a certificate and private key match
func (v *CertificateValidator) ValidateCertificateKeyPair(certPEM, keyPEM string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []string{},
	}

	// Validate certificate
	certResult, err := v.ValidateCertificatePEM(certPEM)
	if err != nil {
		return nil, err
	}
	if !certResult.Valid {
		result.Valid = false
		result.Errors = append(result.Errors, certResult.Errors...)
		return result, nil
	}
	result.Certificate = certResult.Certificate
	result.Warnings = append(result.Warnings, certResult.Warnings...)

	// Validate private key
	keyResult, err := v.ValidatePrivateKeyPEM(keyPEM)
	if err != nil {
		return nil, err
	}
	if !keyResult.Valid {
		result.Valid = false
		result.Errors = append(result.Errors, keyResult.Errors...)
		return result, nil
	}
	result.PrivateKey = keyResult.PrivateKey
	result.Warnings = append(result.Warnings, keyResult.Warnings...)

	// Check if certificate and key match
	if err := v.checkKeyPairMatch(certResult.Certificate, keyResult.PrivateKey); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "certificate_key_pair",
			Message: fmt.Sprintf("certificate and private key do not match: %v", err),
		})
	}

	return result, nil
}

// validateCertificateProperties validates certificate properties
func (v *CertificateValidator) validateCertificateProperties(cert *x509.Certificate, result *ValidationResult) {
	now := time.Now()

	// Check if certificate is expired
	if now.After(cert.NotAfter) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "certificate",
			Message: fmt.Sprintf("certificate has expired (valid until: %s)", cert.NotAfter.Format(time.RFC3339)),
		})
	}

	// Check if certificate is not yet valid
	if now.Before(cert.NotBefore) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "certificate",
			Message: fmt.Sprintf("certificate is not yet valid (valid from: %s)", cert.NotBefore.Format(time.RFC3339)),
		})
	}

	// Warn if certificate is expiring soon (within 30 days)
	expiresIn := cert.NotAfter.Sub(now)
	if expiresIn > 0 && expiresIn < 30*24*time.Hour {
		daysLeft := int(expiresIn.Hours() / 24)
		result.Warnings = append(result.Warnings, fmt.Sprintf("certificate will expire in %d days", daysLeft))
	}

	// Check if certificate has a subject
	if cert.Subject.CommonName == "" && len(cert.DNSNames) == 0 && len(cert.IPAddresses) == 0 {
		result.Warnings = append(result.Warnings, "certificate has no subject common name, DNS names, or IP addresses")
	}

	// Check key usage
	if cert.KeyUsage == 0 {
		result.Warnings = append(result.Warnings, "certificate has no key usage set")
	}

	// Check if it's a CA certificate being used as a server certificate
	if cert.IsCA {
		result.Warnings = append(result.Warnings, "certificate appears to be a CA certificate - it should not be used as a server certificate")
	}
}

// validatePrivateKeyProperties validates private key properties
func (v *CertificateValidator) validatePrivateKeyProperties(key crypto.PrivateKey, result *ValidationResult) {
	switch k := key.(type) {
	case *rsa.PrivateKey:
		// Check RSA key size
		keySize := k.N.BitLen()
		if keySize < 2048 {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "private_key",
				Message: fmt.Sprintf("RSA key size too small: %d bits (minimum: 2048 bits)", keySize),
			})
		} else if keySize < 3072 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("RSA key size is %d bits - consider using at least 3072 bits for better security", keySize))
		}

	case *ecdsa.PrivateKey:
		// Check EC curve
		curveName := k.Curve.Params().Name
		weakCurves := []string{"P-224", "secp256k1"}
		for _, weak := range weakCurves {
			if curveName == weak {
				result.Warnings = append(result.Warnings, fmt.Sprintf("EC curve %s is not recommended - consider using P-256, P-384, or P-521", curveName))
				break
			}
		}
	}
}

// checkKeyPairMatch verifies that the private key matches the certificate's public key
func (v *CertificateValidator) checkKeyPairMatch(cert *x509.Certificate, privateKey crypto.PrivateKey) error {
	// Create a test message
	testMessage := []byte("certificate-key-pair-validation-test")
	hash := sha256.Sum256(testMessage)

	switch pub := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		// Check if private key is RSA
		priv, ok := privateKey.(*rsa.PrivateKey)
		if !ok {
			return fmt.Errorf("certificate has RSA public key but private key is not RSA")
		}

		// Check if public keys match by comparing modulus
		if pub.N.Cmp(priv.N) != 0 {
			return fmt.Errorf("RSA public key modulus does not match private key modulus")
		}

	case *ecdsa.PublicKey:
		// Check if private key is ECDSA
		priv, ok := privateKey.(*ecdsa.PrivateKey)
		if !ok {
			return fmt.Errorf("certificate has ECDSA public key but private key is not ECDSA")
		}

		// Check if public keys match by comparing curve and coordinates
		if pub.Curve != priv.Curve {
			return fmt.Errorf("ECDSA curve mismatch")
		}
		if pub.X.Cmp(priv.X) != 0 || pub.Y.Cmp(priv.Y) != 0 {
			return fmt.Errorf("ECDSA public key coordinates do not match private key")
		}

	default:
		return fmt.Errorf("unsupported public key type: %T", pub)
	}

	// Additional verification: try to sign and verify
	_ = hash // We've already compared the keys directly, which is more reliable

	return nil
}

// ValidateForDomain validates that a certificate is suitable for a specific domain
func (v *CertificateValidator) ValidateForDomain(certPEM string, domain string) (*ValidationResult, error) {
	result, err := v.ValidateCertificatePEM(certPEM)
	if err != nil {
		return nil, err
	}

	if !result.Valid || result.Certificate == nil {
		return result, nil
	}

	cert := result.Certificate

	// Check if domain matches certificate
	matched := false

	// Check Subject Common Name
	if cert.Subject.CommonName == domain || matchWildcard(cert.Subject.CommonName, domain) {
		matched = true
	}

	// Check Subject Alternative Names (SANs)
	if !matched {
		for _, san := range cert.DNSNames {
			if san == domain || matchWildcard(san, domain) {
				matched = true
				break
			}
		}
	}

	if !matched {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "certificate",
			Message: fmt.Sprintf("certificate is not valid for domain '%s' (CN: %s, SANs: %v)", domain, cert.Subject.CommonName, cert.DNSNames),
		})
	}

	return result, nil
}

// matchWildcard checks if a wildcard certificate matches a domain
func matchWildcard(pattern, domain string) bool {
	if !strings.HasPrefix(pattern, "*.") {
		return pattern == domain
	}

	// Wildcard certificate (e.g., *.example.com)
	wildcardDomain := pattern[2:] // Remove "*."

	// Check if domain ends with the wildcard domain
	if !strings.HasSuffix(domain, wildcardDomain) {
		return false
	}

	// Check if the domain has exactly one more label than the wildcard domain
	domainLabels := strings.Split(domain, ".")
	wildcardLabels := strings.Split(wildcardDomain, ".")

	return len(domainLabels) == len(wildcardLabels)+1
}
