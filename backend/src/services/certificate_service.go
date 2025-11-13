package services

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"time"

	"github.com/domain-manager/backend/src/db"
	"github.com/domain-manager/backend/src/k8s"
	"github.com/domain-manager/backend/src/models"
	"github.com/domain-manager/backend/src/repositories"
	"github.com/domain-manager/backend/src/services/certificate"
	"github.com/domain-manager/backend/src/services/letsencrypt"
)

// CertificateService handles certificate business logic
type CertificateService struct {
	certRepo   *repositories.CertificateRepository
	domainRepo *repositories.DomainRepository
	validator  *certificate.CertificateValidator
	encryptor  *certificate.Encryptor
}

// NewCertificateService creates a new certificate service
func NewCertificateService(
	certRepo *repositories.CertificateRepository,
	domainRepo *repositories.DomainRepository,
) *CertificateService {
	encryptor, err := certificate.NewEncryptor()
	if err != nil {
		log.Printf("⚠️  Failed to initialize encryptor: %v (private keys will NOT be encrypted)", err)
		encryptor = nil
	}

	return &CertificateService{
		certRepo:   certRepo,
		domainRepo: domainRepo,
		validator:  certificate.NewValidator(),
		encryptor:  encryptor,
	}
}

// UploadCertificate uploads a certificate for a domain
func (s *CertificateService) UploadCertificate(
	domainName string,
	certPEM string,
	keyPEM string,
) (*models.Certificate, error) {
	// 使用新的驗證服務進行完整驗證
	validationResult, err := s.validator.ValidateCertificateKeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// 檢查驗證結果
	if !validationResult.Valid {
		errorMsg := "certificate validation failed:"
		for _, verr := range validationResult.Errors {
			errorMsg += fmt.Sprintf("\n  - %s", verr.Error())
		}
		return nil, fmt.Errorf(errorMsg)
	}

	// 記錄警告（如果有的話）
	if len(validationResult.Warnings) > 0 {
		for _, warning := range validationResult.Warnings {
			log.Printf("⚠️  Certificate warning: %s", warning)
		}
	}

	// 驗證憑證是否適用於指定的域名
	domainValidation, err := s.validator.ValidateForDomain(certPEM, domainName)
	if err != nil {
		return nil, fmt.Errorf("domain validation error: %w", err)
	}
	if !domainValidation.Valid {
		errorMsg := "certificate is not valid for domain:"
		for _, verr := range domainValidation.Errors {
			errorMsg += fmt.Sprintf("\n  - %s", verr.Error())
		}
		return nil, fmt.Errorf(errorMsg)
	}

	// 使用驗證後的憑證物件
	parsedCert := validationResult.Certificate

	// 檢查域名是否存在
	domain, err := s.domainRepo.GetByName(domainName)
	if err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}

	// 加密私鑰（如果有 encryptor）
	privateKeyToStore := keyPEM
	if s.encryptor != nil {
		encryptedKey, err := s.encryptor.EncryptPrivateKey(keyPEM)
		if err != nil {
			log.Printf("⚠️  Failed to encrypt private key: %v (storing unencrypted)", err)
		} else {
			privateKeyToStore = encryptedKey
			log.Printf("✅ Successfully encrypted private key for domain %s", domainName)
		}
	}

	// 建立憑證記錄
	cert := &models.Certificate{
		DomainName:         domainName,
		Source:             "manual",
		CertificatePEM:     certPEM,
		PrivateKeyPEM:      privateKeyToStore,
		Issuer:             parsedCert.Issuer.CommonName,
		ValidFrom:          parsedCert.NotBefore,
		ValidUntil:         parsedCert.NotAfter,
		K8sSecretName:      fmt.Sprintf("cert-%s", domainName),
		K8sSecretNamespace: domain.TargetNamespace,
		AutoRenew:          false,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// 更新狀態
	cert.UpdateStatus()

	// 儲存到資料庫
	if err := s.certRepo.Create(cert); err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// 更新域名的憑證關聯
	certID := int64(cert.ID)
	domain.CertificateID = &certID
	if err := s.domainRepo.Update(domain); err != nil {
		log.Printf("⚠️  Failed to update domain certificate_id: %v", err)
	} else {
		log.Printf("✅ Associated certificate ID %d with domain %s", cert.ID, domain.DomainName)
	}

	// 建立 K8s Secret
	go s.createSecretForCertificate(cert, domain)

	return cert, nil
}

// ListCertificates lists certificates with pagination (includes both manual and cert-manager managed)
func (s *CertificateService) ListCertificates(limit, offset int) ([]*models.Certificate, int, error) {
	if limit == 0 {
		limit = 50
	}

	// Get manual certificates from database
	dbCerts, err := s.certRepo.List(limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list certificates: %w", err)
	}

	// Get cert-manager managed certificates from Kubernetes
	certManagerCerts, err := s.listCertManagerCertificates()
	if err != nil {
		log.Printf("Warning: Failed to list cert-manager certificates: %v", err)
		// Continue with just database certificates
	}

	// Merge both sources
	allCerts := append(dbCerts, certManagerCerts...)

	// Count total
	dbCount, err := s.certRepo.Count()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count certificates: %w", err)
	}
	totalCount := dbCount + len(certManagerCerts)

	return allCerts, totalCount, nil
}

// listCertManagerCertificates retrieves certificates managed by cert-manager from Kubernetes Secrets
func (s *CertificateService) listCertManagerCertificates() ([]*models.Certificate, error) {
	var certs []*models.Certificate

	// Get all domains with SSL mode = "auto"
	rows, err := db.DB.Query(`
		SELECT id, domain_name, target_namespace, status
		FROM domains
		WHERE ssl_mode = 'auto' AND enabled = 1
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query auto-ssl domains: %w", err)
	}
	defer rows.Close()

	secretMgr := k8s.NewSecretManager()

	for rows.Next() {
		var domainID int64
		var domainName, namespace, status string
		if err := rows.Scan(&domainID, &domainName, &namespace, &status); err != nil {
			log.Printf("Failed to scan domain: %v", err)
			continue
		}

		// Check if TLS secret exists for this domain
		secretName := fmt.Sprintf("domain-%d-tls", domainID)
		secret, err := secretMgr.GetSecret(namespace, secretName)
		if err != nil {
			// Secret doesn't exist yet, skip
			continue
		}

		// Parse certificate from secret
		certPEM, ok := secret.Data["tls.crt"]
		if !ok {
			continue
		}

		block, _ := pem.Decode(certPEM)
		if block == nil {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			log.Printf("Failed to parse certificate for %s: %v", domainName, err)
			continue
		}

		// Create Certificate model
		certModel := &models.Certificate{
			ID:              int(domainID * -1), // Use negative ID to distinguish from DB certs
			DomainName:      domainName,
			Source:          "cert-manager",
			CertificatePEM:  string(certPEM),
			PrivateKeyPEM:   "", // Don't expose private key in list
			Issuer:          cert.Issuer.CommonName,
			ValidFrom:       cert.NotBefore,
			ValidUntil:      cert.NotAfter,
			Status:          s.determineCertStatus(cert.NotAfter),
			K8sSecretName:   secretName,
			K8sSecretNamespace: namespace,
			AutoRenew:       true,
			CreatedAt:       secret.CreationTimestamp.Time,
			UpdatedAt:       secret.CreationTimestamp.Time,
		}

		certs = append(certs, certModel)
	}

	return certs, nil
}

// determineCertStatus determines certificate status based on expiry date
func (s *CertificateService) determineCertStatus(validUntil time.Time) string {
	now := time.Now()
	daysUntilExpiry := int(validUntil.Sub(now).Hours() / 24)

	if daysUntilExpiry < 0 {
		return "expired"
	} else if daysUntilExpiry < 30 {
		return "expiring"
	}
	return "valid"
}

// GetCertificateByID retrieves a certificate by ID
func (s *CertificateService) GetCertificateByID(id int) (*models.Certificate, error) {
	cert, err := s.certRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return cert, nil
}

// GetCertificateByDomain retrieves the latest certificate for a domain
func (s *CertificateService) GetCertificateByDomain(domainName string) (*models.Certificate, error) {
	cert, err := s.certRepo.GetByDomain(domainName)
	if err != nil {
		return nil, err
	}
	return cert, nil
}

// DeleteCertificate deletes a certificate
func (s *CertificateService) DeleteCertificate(id int) error {
	cert, err := s.certRepo.GetByID(id)
	if err != nil {
		return err
	}

	// 刪除 K8s Secret
	go s.deleteSecretForCertificate(cert)

	// 從資料庫刪除
	if err := s.certRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete certificate: %w", err)
	}

	return nil
}

// GetExpiringCertificates retrieves certificates expiring within specified days
func (s *CertificateService) GetExpiringCertificates(days int) ([]*models.Certificate, error) {
	certs, err := s.certRepo.GetExpiring(days)
	if err != nil {
		return nil, fmt.Errorf("failed to get expiring certificates: %w", err)
	}
	return certs, nil
}

// UpdateCertificateStatus updates certificate status based on validity
func (s *CertificateService) UpdateCertificateStatus(id int) error {
	cert, err := s.certRepo.GetByID(id)
	if err != nil {
		return err
	}

	cert.UpdateStatus()

	if err := s.certRepo.Update(cert); err != nil {
		return fmt.Errorf("failed to update certificate: %w", err)
	}

	return nil
}

// createSecretForCertificate 為憑證建立 K8s Secret
func (s *CertificateService) createSecretForCertificate(cert *models.Certificate, domain *models.Domain) {
	secretMgr := k8s.NewSecretManager()

	// 解密私鑰（如果有 encryptor）
	privateKeyPEM := cert.PrivateKeyPEM
	if s.encryptor != nil {
		decryptedKey, err := s.encryptor.DecryptPrivateKey(cert.PrivateKeyPEM)
		if err != nil {
			log.Printf("⚠️  Failed to decrypt private key for certificate %d: %v (using stored value)", cert.ID, err)
		} else {
			privateKeyPEM = decryptedKey
		}
	}

	// 建立 TLS Secret
	_, err := secretMgr.CreateTLSSecret(
		cert.K8sSecretNamespace,
		cert.K8sSecretName,
		cert.CertificatePEM,
		privateKeyPEM,
	)
	if err != nil {
		log.Printf("❌ Failed to create K8s Secret for certificate %d: %v", cert.ID, err)
		return
	}

	log.Printf("✅ Successfully created K8s Secret for certificate %d", cert.ID)

	// 更新 Domain 的 Ingress 以使用此憑證
	go s.updateDomainIngressWithCertificate(domain, cert)
}

// deleteSecretForCertificate 刪除憑證的 K8s Secret
func (s *CertificateService) deleteSecretForCertificate(cert *models.Certificate) {
	secretMgr := k8s.NewSecretManager()

	err := secretMgr.DeleteSecret(cert.K8sSecretNamespace, cert.K8sSecretName)
	if err != nil {
		log.Printf("❌ Failed to delete K8s Secret for certificate %d: %v", cert.ID, err)
		return
	}

	log.Printf("✅ Successfully deleted K8s Secret for certificate %d", cert.ID)
}

// updateDomainIngressWithCertificate 更新 Domain 的 Ingress 以使用憑證
func (s *CertificateService) updateDomainIngressWithCertificate(domain *models.Domain, cert *models.Certificate) {
	ingressMgr := k8s.NewIngressManager()

	ingressClassName := "nginx"
	cfg := &k8s.IngressConfig{
		Name:             fmt.Sprintf("domain-%d", domain.ID),
		Namespace:        domain.TargetNamespace,
		Host:             domain.DomainName,
		ServiceName:      domain.TargetService,
		ServicePort:      domain.TargetPort,
		TLSSecretName:    cert.K8sSecretName,
		IngressClassName: &ingressClassName,
	}

	_, err := ingressMgr.UpdateIngress(cfg)
	if err != nil {
		log.Printf("❌ Failed to update Ingress with certificate for domain %s: %v", domain.DomainName, err)
		return
	}

	log.Printf("✅ Successfully updated Ingress with certificate for domain %s", domain.DomainName)
}

// IsWildcardCertificate checks if a certificate is a wildcard certificate
func (s *CertificateService) IsWildcardCertificate(certPEM string) (bool, error) {
	cert, err := models.ParseCertificatePEM(certPEM)
	if err != nil {
		return false, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Check if any DNS name starts with *.
	for _, dnsName := range cert.DNSNames {
		if len(dnsName) > 2 && dnsName[0] == '*' && dnsName[1] == '.' {
			return true, nil
		}
	}

	// Check CommonName as fallback
	if len(cert.Subject.CommonName) > 2 && cert.Subject.CommonName[0] == '*' && cert.Subject.CommonName[1] == '.' {
		return true, nil
	}

	return false, nil
}

// GetWildcardDomains returns all wildcard domains covered by a certificate
func (s *CertificateService) GetWildcardDomains(certPEM string) ([]string, error) {
	cert, err := models.ParseCertificatePEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	var wildcardDomains []string
	seen := make(map[string]bool)

	// Extract wildcard domains from SANs
	for _, dnsName := range cert.DNSNames {
		if len(dnsName) > 2 && dnsName[0] == '*' && dnsName[1] == '.' {
			if !seen[dnsName] {
				wildcardDomains = append(wildcardDomains, dnsName)
				seen[dnsName] = true
			}
		}
	}

	// Check CommonName as fallback
	cn := cert.Subject.CommonName
	if len(cn) > 2 && cn[0] == '*' && cn[1] == '.' {
		if !seen[cn] {
			wildcardDomains = append(wildcardDomains, cn)
		}
	}

	return wildcardDomains, nil
}

// CanCertificateCoverDomain checks if a certificate can cover a specific domain
// Supports both exact matches and wildcard certificates
func (s *CertificateService) CanCertificateCoverDomain(certPEM string, domainName string) (bool, error) {
	cert, err := models.ParseCertificatePEM(certPEM)
	if err != nil {
		return false, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Check exact match first
	for _, dnsName := range cert.DNSNames {
		if dnsName == domainName {
			return true, nil
		}
	}

	// Check CommonName
	if cert.Subject.CommonName == domainName {
		return true, nil
	}

	// Check wildcard match
	for _, dnsName := range cert.DNSNames {
		if s.matchesWildcard(dnsName, domainName) {
			return true, nil
		}
	}

	// Check CommonName wildcard
	if s.matchesWildcard(cert.Subject.CommonName, domainName) {
		return true, nil
	}

	return false, nil
}

// matchesWildcard checks if a wildcard pattern matches a domain
// e.g., *.example.com matches api.example.com but not example.com or sub.api.example.com
func (s *CertificateService) matchesWildcard(pattern string, domain string) bool {
	// Pattern must start with *.
	if len(pattern) < 3 || pattern[0] != '*' || pattern[1] != '.' {
		return false
	}

	// Extract root domain from pattern (remove *.)
	rootDomain := pattern[2:]

	// Domain must end with the root domain
	if !endsWith(domain, rootDomain) {
		return false
	}

	// Ensure domain is exactly one level deeper than root
	// e.g., api.example.com is valid for *.example.com
	// but sub.api.example.com is not valid for *.example.com
	prefix := domain[:len(domain)-len(rootDomain)]
	if len(prefix) == 0 {
		// Domain is the root domain itself (no subdomain)
		return false
	}

	// Remove trailing dot from prefix
	if prefix[len(prefix)-1] == '.' {
		prefix = prefix[:len(prefix)-1]
	}

	// Check that prefix doesn't contain dots (only one level)
	for i := 0; i < len(prefix); i++ {
		if prefix[i] == '.' {
			return false // Multiple levels, doesn't match
		}
	}

	return true
}

// endsWith checks if s ends with suffix
func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

// FindApplicableWildcardCertificate finds a wildcard certificate that can cover the given domain
func (s *CertificateService) FindApplicableWildcardCertificate(domainName string) (*models.Certificate, error) {
	// Get all certificates
	certs, err := s.certRepo.List(1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list certificates: %w", err)
	}

	// Check each certificate
	for _, cert := range certs {
		canCover, err := s.CanCertificateCoverDomain(cert.CertificatePEM, domainName)
		if err != nil {
			log.Printf("⚠️  Failed to check certificate %d: %v", cert.ID, err)
			continue
		}

		if canCover {
			// Check if it's a wildcard certificate
			isWildcard, err := s.IsWildcardCertificate(cert.CertificatePEM)
			if err != nil {
				log.Printf("⚠️  Failed to check if certificate %d is wildcard: %v", cert.ID, err)
				continue
			}

			if isWildcard {
				return cert, nil
			}
		}
	}

	return nil, nil // No applicable wildcard certificate found
}

// ObtainLetsEncryptCertificate obtains a certificate from Let's Encrypt for a domain
func (s *CertificateService) ObtainLetsEncryptCertificate(domainName string) (*models.Certificate, error) {
	// Get domain from database
	domain, err := s.domainRepo.GetByName(domainName)
	if err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}

	// Validate domain
	if err := letsencrypt.ValidateDomain(domainName); err != nil {
		return nil, fmt.Errorf("invalid domain: %w", err)
	}

	// Get Let's Encrypt client
	leClient, err := letsencrypt.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Let's Encrypt client not initialized: %w", err)
	}

	// Obtain certificate
	log.Printf("Requesting Let's Encrypt certificate for domain: %s", domainName)
	result, err := leClient.ObtainCertificateWithRetry(&letsencrypt.ObtainOptions{
		Domains:  []string{domainName},
		HTTPPort: 80,
		Timeout:  30 * time.Second,
	}, 3) // Retry up to 3 times

	if err != nil {
		return nil, fmt.Errorf("failed to obtain certificate: %w", err)
	}

	// Save certificate to database
	cert := &models.Certificate{
		DomainName:         domainName,
		Source:             "letsencrypt",
		CertificatePEM:     string(result.Certificate),
		PrivateKeyPEM:      string(result.PrivateKey),
		Issuer:             "Let's Encrypt",
		ValidFrom:          result.NotBefore,
		ValidUntil:         result.NotAfter,
		K8sSecretName:      fmt.Sprintf("cert-%s", domainName),
		K8sSecretNamespace: domain.TargetNamespace,
		AutoRenew:          true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Encrypt private key if encryptor is available
	if s.encryptor != nil {
		encryptedKey, err := s.encryptor.EncryptPrivateKey(cert.PrivateKeyPEM)
		if err != nil {
			log.Printf("⚠️  Failed to encrypt private key: %v (storing unencrypted)", err)
		} else {
			cert.PrivateKeyPEM = encryptedKey
			log.Printf("✅ Successfully encrypted private key for domain %s", domainName)
		}
	}

	cert.UpdateStatus()

	if err := s.certRepo.Create(cert); err != nil {
		return nil, fmt.Errorf("failed to save certificate: %w", err)
	}

	// Update domain certificate association
	certID := int64(cert.ID)
	domain.CertificateID = &certID
	if err := s.domainRepo.Update(domain); err != nil {
		log.Printf("⚠️  Failed to update domain certificate_id: %v", err)
	}

	// Create K8s Secret asynchronously
	go s.createSecretForCertificate(cert, domain)

	log.Printf("✅ Successfully obtained Let's Encrypt certificate for %s", domainName)
	return cert, nil
}

// RenewCertificate renews an existing certificate
func (s *CertificateService) RenewCertificate(id int) (*models.Certificate, error) {
	// Get existing certificate
	cert, err := s.certRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("certificate not found: %w", err)
	}

	// Only renew Let's Encrypt certificates
	if cert.Source != "letsencrypt" {
		return nil, fmt.Errorf("can only renew Let's Encrypt certificates")
	}

	// Get domain
	domain, err := s.domainRepo.GetByName(cert.DomainName)
	if err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}

	// Decrypt private key if needed
	privateKeyPEM := cert.PrivateKeyPEM
	if s.encryptor != nil {
		decryptedKey, err := s.encryptor.DecryptPrivateKey(cert.PrivateKeyPEM)
		if err != nil {
			log.Printf("⚠️  Failed to decrypt private key: %v", err)
		} else {
			privateKeyPEM = decryptedKey
		}
	}

	// Get Let's Encrypt client
	leClient, err := letsencrypt.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Let's Encrypt client not initialized: %w", err)
	}

	// Renew certificate
	log.Printf("Renewing certificate for domain: %s", cert.DomainName)
	result, err := leClient.RenewCertificate(&letsencrypt.RenewalOptions{
		CertPEM:    []byte(cert.CertificatePEM),
		PrivateKey: []byte(privateKeyPEM),
		Domains:    []string{cert.DomainName},
		HTTPPort:   80,
		Timeout:    30 * time.Second,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to renew certificate: %w", err)
	}

	if !result.Renewed {
		log.Printf("Certificate for %s doesn't need renewal yet: %s", cert.DomainName, result.Reason)
		return cert, nil
	}

	// Update certificate in database
	cert.CertificatePEM = string(result.NewCertificate)
	cert.PrivateKeyPEM = string(result.NewPrivateKey)
	cert.ValidUntil = result.ExpiresAt
	cert.UpdatedAt = time.Now()

	// Encrypt private key if encryptor is available
	if s.encryptor != nil {
		encryptedKey, err := s.encryptor.EncryptPrivateKey(cert.PrivateKeyPEM)
		if err != nil {
			log.Printf("⚠️  Failed to encrypt private key: %v (storing unencrypted)", err)
		} else {
			cert.PrivateKeyPEM = encryptedKey
		}
	}

	cert.UpdateStatus()

	if err := s.certRepo.Update(cert); err != nil {
		return nil, fmt.Errorf("failed to update certificate: %w", err)
	}

	// Update K8s Secret asynchronously
	go s.createSecretForCertificate(cert, domain)

	log.Printf("✅ Successfully renewed certificate for %s", cert.DomainName)
	return cert, nil
}

// AutoRenewCertificates checks and renews certificates that are expiring soon
func (s *CertificateService) AutoRenewCertificates(daysBeforeExpiry int) error {
	if daysBeforeExpiry == 0 {
		daysBeforeExpiry = 30
	}

	// Get expiring certificates
	certs, err := s.certRepo.GetExpiring(daysBeforeExpiry)
	if err != nil {
		return fmt.Errorf("failed to get expiring certificates: %w", err)
	}

	log.Printf("Found %d certificate(s) expiring within %d days", len(certs), daysBeforeExpiry)

	for _, cert := range certs {
		// Only auto-renew if enabled
		if !cert.AutoRenew {
			log.Printf("Skipping auto-renewal for %s (auto_renew disabled)", cert.DomainName)
			continue
		}

		// Only renew Let's Encrypt certificates
		if cert.Source != "letsencrypt" {
			log.Printf("Skipping auto-renewal for %s (not a Let's Encrypt certificate)", cert.DomainName)
			continue
		}

		// Renew certificate
		log.Printf("Auto-renewing certificate for %s", cert.DomainName)
		_, err := s.RenewCertificate(cert.ID)
		if err != nil {
			log.Printf("❌ Failed to auto-renew certificate for %s: %v", cert.DomainName, err)
			continue
		}

		log.Printf("✅ Successfully auto-renewed certificate for %s", cert.DomainName)
	}

	return nil
}
