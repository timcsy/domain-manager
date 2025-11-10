package services

import (
	"fmt"
	"log"

	"github.com/domain-manager/backend/src/k8s"
	"github.com/domain-manager/backend/src/models"
	"github.com/domain-manager/backend/src/repositories"
)

// DomainService handles domain business logic
type DomainService struct {
	domainRepo *repositories.DomainRepository
	certRepo   *repositories.CertificateRepository
}

// NewDomainService creates a new domain service
func NewDomainService(domainRepo *repositories.DomainRepository) *DomainService {
	return &DomainService{
		domainRepo: domainRepo,
		certRepo:   nil, // Will be set later if needed
	}
}

// SetCertificateRepository sets the certificate repository
func (s *DomainService) SetCertificateRepository(certRepo *repositories.CertificateRepository) {
	s.certRepo = certRepo
}

// ListDomains retrieves a list of domains
func (s *DomainService) ListDomains(filter models.DomainFilter) ([]*models.Domain, int, error) {
	// Set default limit if not specified
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	domains, err := s.domainRepo.List(filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list domains: %w", err)
	}

	count, err := s.domainRepo.Count(filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count domains: %w", err)
	}

	return domains, count, nil
}

// GetDomainByID retrieves a domain by ID
func (s *DomainService) GetDomainByID(id int64) (*models.Domain, error) {
	domain, err := s.domainRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return domain, nil
}

// CreateDomain creates a new domain
func (s *DomainService) CreateDomain(req *models.DomainCreateRequest) (*models.Domain, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if domain already exists
	existing, err := s.domainRepo.GetByName(req.DomainName)
	if err == nil && existing != nil {
		return nil, models.ErrDomainExists
	}

	// Create domain
	domain := &models.Domain{
		DomainName:      req.DomainName,
		TargetService:   req.TargetService,
		TargetNamespace: req.TargetNamespace,
		TargetPort:      req.TargetPort,
		SSLMode:         req.SSLMode,
	}

	if err := s.domainRepo.Create(domain); err != nil {
		return nil, fmt.Errorf("failed to create domain: %w", err)
	}

	// Create Kubernetes Ingress resource
	go s.createIngressForDomain(domain)

	return domain, nil
}

// UpdateDomain updates an existing domain
func (s *DomainService) UpdateDomain(id int64, req *models.DomainUpdateRequest) (*models.Domain, error) {
	// Get existing domain
	domain, err := s.domainRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.TargetService != nil {
		domain.TargetService = *req.TargetService
	}
	if req.TargetNamespace != nil {
		domain.TargetNamespace = *req.TargetNamespace
	}
	if req.TargetPort != nil {
		domain.TargetPort = *req.TargetPort
	}
	if req.SSLMode != nil {
		domain.SSLMode = *req.SSLMode
	}
	if req.Enabled != nil {
		domain.Enabled = *req.Enabled
	}

	// Save changes
	if err := s.domainRepo.Update(domain); err != nil {
		return nil, fmt.Errorf("failed to update domain: %w", err)
	}

	// Update Kubernetes Ingress resource
	go s.updateIngressForDomain(domain)

	return domain, nil
}

// DeleteDomain deletes a domain
func (s *DomainService) DeleteDomain(id int64, hard bool) error {
	// Get domain
	domain, err := s.domainRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Delete Kubernetes Ingress resource
	go s.deleteIngressForDomain(domain)

	if hard {
		// Hard delete
		if err := s.domainRepo.Delete(id); err != nil {
			return fmt.Errorf("failed to delete domain: %w", err)
		}
	} else {
		// Soft delete
		if err := s.domainRepo.SoftDelete(id); err != nil {
			return fmt.Errorf("failed to soft delete domain: %w", err)
		}
	}

	_ = domain // Use domain for K8s operations
	return nil
}

// createIngressForDomain 為域名建立 Kubernetes Ingress 資源
func (s *DomainService) createIngressForDomain(domain *models.Domain) {
	ingressMgr := k8s.NewIngressManager()

	// 準備 Ingress 配置
	cfg := &k8s.IngressConfig{
		Name:          fmt.Sprintf("domain-%d", domain.ID),
		Namespace:     domain.TargetNamespace,
		Host:          domain.DomainName,
		ServiceName:   domain.TargetService,
		ServicePort:   domain.TargetPort,
		TLSSecretName: "", // 將在建立憑證時設定
	}

	// 建立 Ingress
	_, err := ingressMgr.CreateIngress(cfg)
	if err != nil {
		log.Printf("❌ Failed to create Ingress for domain %s: %v", domain.DomainName, err)
		domain.Status = "error"
		s.domainRepo.Update(domain)
		return
	}

	// 更新域名狀態
	domain.Status = "active"
	s.domainRepo.Update(domain)
	log.Printf("✅ Successfully created Ingress for domain %s", domain.DomainName)
}

// updateIngressForDomain 更新域名的 Kubernetes Ingress 資源
func (s *DomainService) updateIngressForDomain(domain *models.Domain) {
	ingressMgr := k8s.NewIngressManager()

	// 準備 Ingress 配置
	cfg := &k8s.IngressConfig{
		Name:          fmt.Sprintf("domain-%d", domain.ID),
		Namespace:     domain.TargetNamespace,
		Host:          domain.DomainName,
		ServiceName:   domain.TargetService,
		ServicePort:   domain.TargetPort,
		TLSSecretName: "", // TODO: 從 certificate 取得
	}

	// 更新 Ingress
	_, err := ingressMgr.UpdateIngress(cfg)
	if err != nil {
		log.Printf("❌ Failed to update Ingress for domain %s: %v", domain.DomainName, err)
		return
	}

	log.Printf("✅ Successfully updated Ingress for domain %s", domain.DomainName)
}

// deleteIngressForDomain 刪除域名的 Kubernetes Ingress 資源
func (s *DomainService) deleteIngressForDomain(domain *models.Domain) {
	ingressMgr := k8s.NewIngressManager()

	ingressName := fmt.Sprintf("domain-%d", domain.ID)
	err := ingressMgr.DeleteIngress(domain.TargetNamespace, ingressName)
	if err != nil {
		log.Printf("❌ Failed to delete Ingress for domain %s: %v", domain.DomainName, err)
		return
	}

	log.Printf("✅ Successfully deleted Ingress for domain %s", domain.DomainName)
}

// GetDomainStatus retrieves the status of a domain
func (s *DomainService) GetDomainStatus(id int64) (map[string]interface{}, error) {
	domain, err := s.domainRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// TODO: Check Kubernetes Ingress status
	// TODO: Check DNS resolution

	// Check certificate status
	certStatus := "none"
	if domain.CertificateID != nil && s.certRepo != nil {
		cert, err := s.certRepo.GetByID(int(*domain.CertificateID))
		if err == nil && cert != nil {
			certStatus = cert.Status
		}
	}

	status := map[string]interface{}{
		"domain":      domain.DomainName,
		"status":      domain.Status,
		"enabled":     domain.Enabled,
		"ingress":     "pending", // TODO: Get actual status
		"certificate": certStatus,
		"dns":         "pending", // TODO: Get actual status
	}

	return status, nil
}
