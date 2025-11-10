package services

import (
	"fmt"
	"log"
	"time"

	"github.com/domain-manager/backend/src/k8s"
	"github.com/domain-manager/backend/src/models"
	"github.com/domain-manager/backend/src/repositories"
)

// CertificateService handles certificate business logic
type CertificateService struct {
	certRepo   *repositories.CertificateRepository
	domainRepo *repositories.DomainRepository
}

// NewCertificateService creates a new certificate service
func NewCertificateService(
	certRepo *repositories.CertificateRepository,
	domainRepo *repositories.DomainRepository,
) *CertificateService {
	return &CertificateService{
		certRepo:   certRepo,
		domainRepo: domainRepo,
	}
}

// UploadCertificate uploads a certificate for a domain
func (s *CertificateService) UploadCertificate(
	domainName string,
	certPEM string,
	keyPEM string,
) (*models.Certificate, error) {
	// 驗證 PEM 格式
	parsedCert, err := models.ParseCertificatePEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("invalid certificate PEM: %w", err)
	}

	// 驗證私鑰
	if err := models.ValidatePrivateKey(keyPEM); err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	// 驗證憑證與私鑰是否匹配
	if err := models.ValidateCertificateKeyPair(certPEM, keyPEM); err != nil {
		return nil, fmt.Errorf("certificate and key do not match: %w", err)
	}

	// 檢查域名是否存在
	domain, err := s.domainRepo.GetByName(domainName)
	if err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}

	// 建立憑證記錄
	cert := &models.Certificate{
		DomainName:         domainName,
		Source:             "manual",
		CertificatePEM:     certPEM,
		PrivateKeyPEM:      keyPEM,
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

// ListCertificates lists certificates with pagination
func (s *CertificateService) ListCertificates(limit, offset int) ([]*models.Certificate, int, error) {
	if limit == 0 {
		limit = 50
	}

	certs, err := s.certRepo.List(limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list certificates: %w", err)
	}

	count, err := s.certRepo.Count()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count certificates: %w", err)
	}

	return certs, count, nil
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

	// 建立 TLS Secret
	_, err := secretMgr.CreateTLSSecret(
		cert.K8sSecretNamespace,
		cert.K8sSecretName,
		cert.CertificatePEM,
		cert.PrivateKeyPEM,
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

	cfg := &k8s.IngressConfig{
		Name:          fmt.Sprintf("domain-%d", domain.ID),
		Namespace:     domain.TargetNamespace,
		Host:          domain.DomainName,
		ServiceName:   domain.TargetService,
		ServicePort:   domain.TargetPort,
		TLSSecretName: cert.K8sSecretName,
	}

	_, err := ingressMgr.UpdateIngress(cfg)
	if err != nil {
		log.Printf("❌ Failed to update Ingress with certificate for domain %s: %v", domain.DomainName, err)
		return
	}

	log.Printf("✅ Successfully updated Ingress with certificate for domain %s", domain.DomainName)
}
