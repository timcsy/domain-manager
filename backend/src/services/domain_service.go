package services

import (
	"fmt"
	"log"
	"strings"

	"encoding/json"

	"github.com/domain-manager/backend/src/k8s"
	"github.com/domain-manager/backend/src/models"
	"github.com/domain-manager/backend/src/repositories"
	"github.com/domain-manager/backend/src/services/subdomain"
)

// DomainService handles domain business logic
type DomainService struct {
	domainRepo         *repositories.DomainRepository
	certRepo           *repositories.CertificateRepository
	settingsService    *SettingsService
	subdomainValidator *subdomain.SubdomainValidator
}

// NewDomainService creates a new domain service
func NewDomainService(domainRepo *repositories.DomainRepository) *DomainService {
	return &DomainService{
		domainRepo:        domainRepo,
		certRepo:          nil, // Will be set later if needed
		subdomainValidator: subdomain.NewValidator(),
	}
}

// SetCertificateRepository sets the certificate repository
func (s *DomainService) SetCertificateRepository(certRepo *repositories.CertificateRepository) {
	s.certRepo = certRepo
}

// SetSettingsService sets the settings service for dynamic configuration
func (s *DomainService) SetSettingsService(settingsService *SettingsService) {
	s.settingsService = settingsService
}

// getCustomAnnotations reads user-defined ingress annotations from system settings
func (s *DomainService) getCustomAnnotations() map[string]string {
	if s.settingsService != nil {
		setting, err := s.settingsService.GetSetting("ingress_annotations")
		if err == nil && setting != nil && setting.Value != "" && setting.Value != "{}" {
			var annotations map[string]string
			if json.Unmarshal([]byte(setting.Value), &annotations) == nil {
				return annotations
			}
		}
	}
	return nil
}

// getIngressAnnotations returns merged annotations for the current controller, SSL state, and extra annotations
func (s *DomainService) getIngressAnnotations(sslEnabled bool, extraAnnotations map[string]string) map[string]string {
	controllerName := s.getIngressClass()
	customAnnotations := s.getCustomAnnotations()

	// Get profile + custom annotations
	result := k8s.GetAnnotationsForController(controllerName, sslEnabled, customAnnotations)

	// Merge extra annotations (e.g., cert-manager) on top
	for k, v := range extraAnnotations {
		result[k] = v
	}

	return result
}

// getIngressClass reads the current default ingress class from system settings
func (s *DomainService) getIngressClass() string {
	if s.settingsService != nil {
		setting, err := s.settingsService.GetSetting("default_ingress_class")
		if err == nil && setting != nil && setting.Value != "" {
			return setting.Value
		}
	}
	return "nginx"
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

	// Check if domain already exists (excluding deleted domains)
	existing, err := s.domainRepo.GetByName(req.DomainName)
	if err == nil && existing != nil && existing.Status != "deleted" {
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
	if req.Status != nil {
		domain.Status = *req.Status
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
	ingressClassName := s.getIngressClass()
	cfg := &k8s.IngressConfig{
		Name:             fmt.Sprintf("domain-%d", domain.ID),
		Namespace:        domain.TargetNamespace,
		Host:             domain.DomainName,
		ServiceName:      domain.TargetService,
		ServicePort:      domain.TargetPort,
		IngressClassName: &ingressClassName,
	}

	// 配置 annotation 和 SSL
	var extraAnnotations map[string]string
	sslEnabled := false
	if domain.SSLMode == "auto" {
		cfg.TLSSecretName = fmt.Sprintf("domain-%d-tls", domain.ID)
		extraAnnotations = map[string]string{
			"cert-manager.io/cluster-issuer": "letsencrypt-prod",
		}
		sslEnabled = true
	}
	cfg.Annotations = s.getIngressAnnotations(sslEnabled, extraAnnotations)

	// 建立 Ingress
	_, err := ingressMgr.CreateIngress(cfg)
	if err != nil {
		log.Printf("❌ Failed to create Ingress for domain %s: %v", domain.DomainName, err)
		domain.Status = "error"
		_ = s.domainRepo.Update(domain)
		return
	}

	// 更新域名狀態
	domain.Status = "active"
	_ = s.domainRepo.Update(domain)
	log.Printf("✅ Successfully created Ingress for domain %s", domain.DomainName)
}

// updateIngressForDomain 更新域名的 Kubernetes Ingress 資源
func (s *DomainService) updateIngressForDomain(domain *models.Domain) {
	ingressMgr := k8s.NewIngressManager()

	// 準備 Ingress 配置
	ingressClassName := s.getIngressClass()
	cfg := &k8s.IngressConfig{
		Name:             fmt.Sprintf("domain-%d", domain.ID),
		Namespace:        domain.TargetNamespace,
		Host:             domain.DomainName,
		ServiceName:      domain.TargetService,
		ServicePort:      domain.TargetPort,
		IngressClassName: &ingressClassName,
	}

	// 配置 annotation 和 SSL
	var extraAnnotations map[string]string
	sslEnabled := false
	if domain.SSLMode == "auto" {
		cfg.TLSSecretName = fmt.Sprintf("domain-%d-tls", domain.ID)
		extraAnnotations = map[string]string{
			"cert-manager.io/cluster-issuer": "letsencrypt-prod",
		}
		sslEnabled = true
	}
	cfg.Annotations = s.getIngressAnnotations(sslEnabled, extraAnnotations)

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

// CheckSubdomainConflict checks for subdomain conflicts with existing domains
func (s *DomainService) CheckSubdomainConflict(domainName string) ([]subdomain.SubdomainConflict, error) {
	// Get all existing domain names
	allDomains, err := s.domainRepo.List(models.DomainFilter{Limit: 1000})
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}

	existingNames := make([]string, len(allDomains))
	for i, d := range allDomains {
		existingNames[i] = d.DomainName
	}

	// Check for conflicts
	conflicts := s.subdomainValidator.CheckSubdomainConflict(domainName, existingNames)
	return conflicts, nil
}

// ValidateSubdomain validates a subdomain format
func (s *DomainService) ValidateSubdomain(domainName string) (*subdomain.ValidationResult, error) {
	return s.subdomainValidator.ValidateSubdomain(domainName)
}

// GroupByRootDomain groups domains by their root domain
// For domains like dns.k8s.tew.tw and dashboard.k8s.tew.tw,
// it finds the common parent (k8s.tew.tw) as the root domain
func (s *DomainService) GroupByRootDomain() (map[string][]*models.Domain, error) {
	allDomains, err := s.domainRepo.List(models.DomainFilter{Limit: 1000})
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}

	// Filter out deleted domains
	var activeDomains []*models.Domain
	for _, domain := range allDomains {
		if domain.Status != "deleted" {
			activeDomains = append(activeDomains, domain)
		}
	}

	if len(activeDomains) == 0 {
		return make(map[string][]*models.Domain), nil
	}

	// Find the longest common suffix for each group
	groups := make(map[string][]*models.Domain)

	for _, domain := range activeDomains {
		// Find the best root domain for this domain
		rootDomain := findBestRootDomain(domain.DomainName, activeDomains)
		groups[rootDomain] = append(groups[rootDomain], domain)
	}

	return groups, nil
}

// findBestRootDomain finds the longest common suffix among domains
// that could be grouped together
func findBestRootDomain(domainName string, allDomains []*models.Domain) string {
	labels := strings.Split(domainName, ".")

	// Try from longest suffix to shortest (but keep at least 2 labels for valid domain)
	for i := 0; i < len(labels)-1; i++ {
		candidate := strings.Join(labels[i:], ".")

		// Check if this candidate is a good grouping point
		// (either it exists as a domain, or multiple domains share it as suffix)
		matchCount := 0
		for _, d := range allDomains {
			if d.DomainName == candidate || strings.HasSuffix(d.DomainName, "."+candidate) {
				matchCount++
			}
		}

		// If multiple domains share this suffix, it's a good root domain
		if matchCount > 1 {
			return candidate
		}
	}

	// Fallback: use the standard root domain (last 2 labels)
	if len(labels) >= 2 {
		return strings.Join(labels[len(labels)-2:], ".")
	}

	return domainName
}

// GetSubdomains retrieves all subdomains for a given root domain
func (s *DomainService) GetSubdomains(rootDomain string) ([]*models.Domain, error) {
	allDomains, err := s.domainRepo.List(models.DomainFilter{Limit: 1000})
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}

	subdomains := []*models.Domain{}
	for _, domain := range allDomains {
		if s.subdomainValidator.IsSubdomain(domain.DomainName) {
			if s.subdomainValidator.GetRootDomain(domain.DomainName) == rootDomain {
				subdomains = append(subdomains, domain)
			}
		}
	}

	return subdomains, nil
}

// BulkOperationResult represents the result of a bulk operation
type BulkOperationResult struct {
	Success      int
	Failed       int
	Total        int
	Errors       []BulkOperationError
	SuccessDomains []*models.Domain
}

// BulkOperationError represents an error in a bulk operation
type BulkOperationError struct {
	DomainName string
	Error      string
}

// BulkCreateDomains creates multiple domains in a single operation
func (s *DomainService) BulkCreateDomains(requests []*models.DomainCreateRequest) *BulkOperationResult {
	result := &BulkOperationResult{
		Total:          len(requests),
		Errors:         []BulkOperationError{},
		SuccessDomains: []*models.Domain{},
	}

	for _, req := range requests {
		domain, err := s.CreateDomain(req)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, BulkOperationError{
				DomainName: req.DomainName,
				Error:      err.Error(),
			})
		} else {
			result.Success++
			result.SuccessDomains = append(result.SuccessDomains, domain)
		}
	}

	log.Printf("✅ Bulk create completed: %d success, %d failed out of %d total", result.Success, result.Failed, result.Total)
	return result
}

// BulkUpdateDomains updates multiple domains in a single operation
func (s *DomainService) BulkUpdateDomains(updates map[int64]*models.DomainUpdateRequest) *BulkOperationResult {
	result := &BulkOperationResult{
		Total:          len(updates),
		Errors:         []BulkOperationError{},
		SuccessDomains: []*models.Domain{},
	}

	for id, req := range updates {
		domain, err := s.UpdateDomain(id, req)
		if err != nil {
			result.Failed++
			// Get domain name for error reporting
			domainName := fmt.Sprintf("ID:%d", id)
			if existingDomain, getErr := s.domainRepo.GetByID(id); getErr == nil && existingDomain != nil {
				domainName = existingDomain.DomainName
			}
			result.Errors = append(result.Errors, BulkOperationError{
				DomainName: domainName,
				Error:      err.Error(),
			})
		} else {
			result.Success++
			result.SuccessDomains = append(result.SuccessDomains, domain)
		}
	}

	log.Printf("✅ Bulk update completed: %d success, %d failed out of %d total", result.Success, result.Failed, result.Total)
	return result
}

// BulkDeleteDomains deletes multiple domains in a single operation
func (s *DomainService) BulkDeleteDomains(ids []int64, hard bool) *BulkOperationResult {
	result := &BulkOperationResult{
		Total:          len(ids),
		Errors:         []BulkOperationError{},
		SuccessDomains: []*models.Domain{},
	}

	for _, id := range ids {
		// Get domain name for result reporting
		domainName := fmt.Sprintf("ID:%d", id)
		existingDomain, getErr := s.domainRepo.GetByID(id)
		if getErr == nil && existingDomain != nil {
			domainName = existingDomain.DomainName
		}

		err := s.DeleteDomain(id, hard)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, BulkOperationError{
				DomainName: domainName,
				Error:      err.Error(),
			})
		} else {
			result.Success++
			if existingDomain != nil {
				result.SuccessDomains = append(result.SuccessDomains, existingDomain)
			}
		}
	}

	log.Printf("✅ Bulk delete completed: %d success, %d failed out of %d total", result.Success, result.Failed, result.Total)
	return result
}

// GetDomainTree builds a tree structure of all domains grouped by root domain
func (s *DomainService) GetDomainTree() ([]*models.DomainTreeNode, error) {
	// Get all domains grouped by root domain
	groups, err := s.GroupByRootDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to group domains: %w", err)
	}

	// Build tree nodes
	var treeNodes []*models.DomainTreeNode
	for rootDomain, domains := range groups {
		// Find the root domain object (if it exists)
		var rootDomainObj *models.Domain
		var subdomains []*models.Domain

		for _, domain := range domains {
			if domain.DomainName == rootDomain {
				rootDomainObj = domain
			} else {
				subdomains = append(subdomains, domain)
			}
		}

		// Build the tree node
		node := &models.DomainTreeNode{
			Domain:     rootDomainObj,
			RootDomain: rootDomain,
			Subdomains: buildSubdomainTree(subdomains, rootDomain),
			Count:      len(domains),
		}

		treeNodes = append(treeNodes, node)
	}

	return treeNodes, nil
}

// buildSubdomainTree recursively builds subdomain tree nodes
func buildSubdomainTree(domains []*models.Domain, parent string) []*models.DomainTreeNode {
	var nodes []*models.DomainTreeNode

	for _, domain := range domains {
		// Check if this domain is a direct subdomain of parent
		if isDirectSubdomain(domain.DomainName, parent) {
			node := &models.DomainTreeNode{
				Domain:     domain,
				RootDomain: parent,
				Subdomains: []*models.DomainTreeNode{},
				Count:      1,
			}

			// Find subdomains of this domain
			var childDomains []*models.Domain
			for _, d := range domains {
				if d.DomainName != domain.DomainName && isSubdomainOf(d.DomainName, domain.DomainName) {
					childDomains = append(childDomains, d)
				}
			}

			if len(childDomains) > 0 {
				node.Subdomains = buildSubdomainTree(childDomains, domain.DomainName)
				node.Count += len(childDomains)
			}

			nodes = append(nodes, node)
		}
	}

	return nodes
}

// isDirectSubdomain checks if domain is a direct subdomain of parent
// e.g., api.example.com is direct subdomain of example.com
// but sub.api.example.com is NOT a direct subdomain of example.com
func isDirectSubdomain(domain, parent string) bool {
	if len(domain) <= len(parent) {
		return false
	}

	// Check if domain ends with parent
	if domain[len(domain)-len(parent):] != parent {
		return false
	}

	// Get the prefix part
	prefix := domain[:len(domain)-len(parent)]

	// Remove trailing dot
	if len(prefix) > 0 && prefix[len(prefix)-1] == '.' {
		prefix = prefix[:len(prefix)-1]
	}

	// Check that prefix doesn't contain dots (only one level)
	for i := 0; i < len(prefix); i++ {
		if prefix[i] == '.' {
			return false
		}
	}

	return len(prefix) > 0
}

// isSubdomainOf checks if domain is a subdomain of parent (direct or nested)
func isSubdomainOf(domain, parent string) bool {
	if len(domain) <= len(parent) {
		return false
	}

	// Check if domain ends with parent
	if domain[len(domain)-len(parent):] != parent {
		return false
	}

	// Check if the character before parent is a dot
	if len(domain) > len(parent) && domain[len(domain)-len(parent)-1] == '.' {
		return true
	}

	return false
}
