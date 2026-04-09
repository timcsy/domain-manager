package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/domain-manager/backend/src/db"
	"github.com/domain-manager/backend/src/k8s"
	"github.com/domain-manager/backend/src/mcp"
	"github.com/domain-manager/backend/src/middleware"
	"github.com/domain-manager/backend/src/models"
	"github.com/domain-manager/backend/src/repositories"
	"github.com/domain-manager/backend/src/services"
	"github.com/domain-manager/backend/src/services/letsencrypt"
	"github.com/domain-manager/backend/src/services/scheduler"
	"github.com/go-chi/chi/v5"
)

var (
	authService        *services.AuthService
	domainService      *services.DomainService
	certificateService *services.CertificateService
	settingsService    *services.SettingsService
	apiKeyService      *services.APIKeyService
	backupService      *services.BackupService
	cloudflareService  *services.CloudflareService
	mcpServer          *mcp.Server
)

// InitializeServices initializes all services
func InitializeServices() {
	adminRepo := repositories.NewAdminAccountRepository(db.DB)
	domainRepo := repositories.NewDomainRepository(db.DB)
	certRepo := repositories.NewCertificateRepository(db.DB)
	settingsRepo := repositories.NewSettingsRepository(db.DB)

	authService = services.NewAuthService(adminRepo)
	domainService = services.NewDomainService(domainRepo)
	domainService.SetCertificateRepository(certRepo)
	certificateService = services.NewCertificateService(certRepo, domainRepo)
	settingsService = services.NewSettingsService(settingsRepo)

	domainService.SetSettingsService(settingsService)
	certificateService.SetSettingsService(settingsService)

	// Apply DEFAULT_INGRESS_CLASS from environment (Helm) if set
	if envIngressClass := os.Getenv("DEFAULT_INGRESS_CLASS"); envIngressClass != "" {
		currentSetting, err := settingsService.GetSetting("default_ingress_class")
		if err != nil || currentSetting == nil || currentSetting.Value == "" || currentSetting.Value == "nginx" {
			_ = settingsService.UpdateSettings(&models.SettingsUpdateRequest{
				Settings: map[string]string{"default_ingress_class": envIngressClass},
			})
			log.Printf("Set default_ingress_class to %s from environment", envIngressClass)
		}
	}

	apiKeyRepo := repositories.NewAPIKeyRepository(db.DB)
	apiKeyService = services.NewAPIKeyService(apiKeyRepo)

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./data/database.db"
	}
	backupService = services.NewBackupService(dbPath)
	cloudflareService = services.NewCloudflareService(settingsService)

	// Apply Cloudflare settings from environment (Helm) if set
	if envCfToken := os.Getenv("CLOUDFLARE_API_TOKEN"); envCfToken != "" {
		if os.Getenv("CLOUDFLARE_ENABLED") == "true" || os.Getenv("CLOUDFLARE_ENABLED") == "1" {
			if err := cloudflareService.SaveToken(envCfToken); err != nil {
				log.Printf("Warning: Failed to set Cloudflare token from environment: %v", err)
			}
		}
	}

	mcpServer = mcp.NewServer(domainService, certificateService, domainRepo, certRepo)

	// Register API key validator for Auth middleware
	middleware.SetAPIKeyValidator(func(rawKey string) (bool, error) {
		_, err := apiKeyService.ValidateKey(rawKey)
		if err != nil {
			return false, err
		}
		return true, nil
	})

	// Initialize Let's Encrypt client
	if err := initializeLetsEncrypt(); err != nil {
		log.Printf("Warning: Let's Encrypt initialization failed: %v", err)
		log.Printf("Let's Encrypt certificate features will be disabled")
	}

	// Seed default admin account if it doesn't exist
	if err := seedDefaultAdmin(adminRepo); err != nil {
		log.Printf("Warning: Failed to seed default admin: %v", err)
	}
}

// seedDefaultAdmin creates a default admin account if none exists
func seedDefaultAdmin(adminRepo *repositories.AdminAccountRepository) error {
	log.Println("Checking for existing admin account...")
	// Check if admin already exists
	_, err := adminRepo.GetByUsername("admin")
	if err == nil {
		log.Println("✅ Admin account already exists")
		// Admin already exists
		return nil
	}
	log.Printf("Admin account not found (error: %v), creating new one...", err)

	// Create default admin account
	log.Println("Creating default admin account (username: admin, password: admin)")
	admin := &models.AdminAccount{
		Username: "admin",
		Email:    "admin@localhost",
	}

	if err := adminRepo.Create(admin, "admin"); err != nil {
		return fmt.Errorf("failed to create admin account: %w", err)
	}

	log.Println("✅ Default admin account created successfully")
	return nil
}

// initializeLetsEncrypt initializes the Let's Encrypt client
func initializeLetsEncrypt() error {
	email := os.Getenv("LETSENCRYPT_EMAIL")
	if email == "" {
		email = "admin@localhost" // Default email
		log.Printf("LETSENCRYPT_EMAIL not set, using default: %s", email)
	}

	accountPath := os.Getenv("LETSENCRYPT_ACCOUNT_PATH")
	if accountPath == "" {
		accountPath = "./data/letsencrypt" // Default path
	}

	staging := os.Getenv("LETSENCRYPT_STAGING") == "true"
	if staging {
		log.Println("Let's Encrypt: Using STAGING environment")
	} else {
		log.Println("Let's Encrypt: Using PRODUCTION environment")
	}

	config := &letsencrypt.Config{
		Email:       email,
		AccountPath: accountPath,
		Staging:     staging,
	}

	if err := letsencrypt.Initialize(config); err != nil {
		return fmt.Errorf("failed to initialize Let's Encrypt: %w", err)
	}

	log.Println("✅ Let's Encrypt client initialized successfully")
	return nil
}

// InitializeScheduler initializes and starts the background task scheduler
func InitializeScheduler() error {
	// Initialize scheduler
	if err := scheduler.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize scheduler: %w", err)
	}

	// Create certificate monitor
	certMonitor := scheduler.NewCertificateMonitor(
		scheduler.DefaultCertificateMonitorConfig(),
		certificateService,
	)

	// Register monitor with scheduler
	certMonitor.RegisterWithScheduler(scheduler.GetScheduler())

	// Start scheduler
	if err := scheduler.StartGlobal(); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	log.Println("✅ Scheduler initialized and started successfully")
	return nil
}

// Auth handlers

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := authService.Login(&req)
	if err != nil {
		if err == models.ErrInvalidCredentials {
			Error(w, http.StatusUnauthorized, "Invalid username or password")
		} else {
			Error(w, http.StatusInternalServerError, "Login failed")
		}
		return
	}

	Success(w, resp, "Login successful")
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token != "" {
		_ = authService.Logout(token)
	}
	Success(w, nil, "Logout successful")
}

// Domain handlers

func HandleListDomains(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	status := r.URL.Query().Get("status")
	parent := r.URL.Query().Get("parent")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	// If parent query parameter is provided, return subdomains
	if parent != "" {
		subdomains, err := domainService.GetSubdomains(parent)
		if err != nil {
			Error(w, http.StatusInternalServerError, "Failed to get subdomains")
			return
		}

		// Apply pagination to subdomains
		total := len(subdomains)
		start := (page - 1) * perPage
		end := start + perPage
		if start > total {
			start = total
		}
		if end > total {
			end = total
		}
		paginatedSubdomains := subdomains[start:end]

		Paginated(w, paginatedSubdomains, page, perPage, total)
		return
	}

	// Otherwise, list all domains with filter
	filter := models.DomainFilter{
		Status: status,
		Limit:  perPage,
		Offset: (page - 1) * perPage,
	}

	domains, total, err := domainService.ListDomains(filter)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to list domains")
		return
	}

	Paginated(w, domains, page, perPage, total)
}

func HandleCreateDomain(w http.ResponseWriter, r *http.Request) {
	var req models.DomainCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	domain, err := domainService.CreateDomain(&req)
	if err != nil {
		if err == models.ErrDomainExists {
			Error(w, http.StatusConflict, "Domain already exists")
		} else {
			Error(w, http.StatusInternalServerError, "Failed to create domain")
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	Success(w, domain, "Domain created successfully")
}

func HandleGetDomain(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid domain ID")
		return
	}

	domain, err := domainService.GetDomainByID(id)
	if err != nil {
		if err == models.ErrDomainNotFound {
			Error(w, http.StatusNotFound, "Domain not found")
		} else {
			Error(w, http.StatusInternalServerError, "Failed to get domain")
		}
		return
	}

	Success(w, domain, "")
}

func HandleUpdateDomain(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid domain ID")
		return
	}

	var req models.DomainUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	domain, err := domainService.UpdateDomain(id, &req)
	if err != nil {
		if err == models.ErrDomainNotFound {
			Error(w, http.StatusNotFound, "Domain not found")
		} else {
			Error(w, http.StatusInternalServerError, "Failed to update domain")
		}
		return
	}

	Success(w, domain, "Domain updated successfully")
}

func HandleDeleteDomain(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid domain ID")
		return
	}

	hard := r.URL.Query().Get("hard") == "true"

	if err := domainService.DeleteDomain(id, hard); err != nil {
		if err == models.ErrDomainNotFound {
			Error(w, http.StatusNotFound, "Domain not found")
		} else {
			Error(w, http.StatusInternalServerError, "Failed to delete domain")
		}
		return
	}

	Success(w, nil, "Domain deleted successfully")
}

func HandleGetDomainStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid domain ID")
		return
	}

	status, err := domainService.GetDomainStatus(id)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to get domain status")
		return
	}

	Success(w, status, "")
}

func HandleGetDomainDiagnostics(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid domain ID")
		return
	}

	diagnosticSvc := services.NewDiagnosticService()
	diagnostics, err := diagnosticSvc.GetDomainDiagnostics(id)
	if err != nil {
		log.Printf("Failed to get domain diagnostics: %v", err)
		Error(w, http.StatusInternalServerError, "Failed to get domain diagnostics")
		return
	}

	Success(w, diagnostics, "Domain diagnostics retrieved successfully")
}

func HandleGetDomainTree(w http.ResponseWriter, r *http.Request) {
	tree, err := domainService.GetDomainTree()
	if err != nil {
		log.Printf("Failed to get domain tree: %v", err)
		Error(w, http.StatusInternalServerError, "Failed to get domain tree")
		return
	}

	response := map[string]interface{}{
		"tree":  tree,
		"count": len(tree),
	}

	Success(w, response, "Domain tree retrieved successfully")
}

func HandleBatchCreateDomains(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Domains []*models.DomainCreateRequest `json:"domains"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.Domains) == 0 {
		Error(w, http.StatusBadRequest, "No domains provided")
		return
	}

	result := domainService.BulkCreateDomains(req.Domains)

	response := map[string]interface{}{
		"total":           result.Total,
		"success":         result.Success,
		"failed":          result.Failed,
		"errors":          result.Errors,
		"success_domains": result.SuccessDomains,
	}

	// Return 207 Multi-Status if there were partial failures
	if result.Failed > 0 && result.Success > 0 {
		w.WriteHeader(http.StatusMultiStatus)
		Success(w, response, "Batch create completed with some failures")
		return
	}

	// Return 400 if all failed
	if result.Failed > 0 && result.Success == 0 {
		w.WriteHeader(http.StatusBadRequest)
		Success(w, response, "All domains failed to create")
		return
	}

	// Return 201 if all succeeded
	w.WriteHeader(http.StatusCreated)
	Success(w, response, "All domains created successfully")
}

func HandleBatchDeleteDomains(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs  []int64 `json:"ids"`
		Hard bool    `json:"hard"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.IDs) == 0 {
		Error(w, http.StatusBadRequest, "No domain IDs provided")
		return
	}

	result := domainService.BulkDeleteDomains(req.IDs, req.Hard)

	response := map[string]interface{}{
		"total":           result.Total,
		"success":         result.Success,
		"failed":          result.Failed,
		"errors":          result.Errors,
		"success_domains": result.SuccessDomains,
	}

	// Return 207 Multi-Status if there were partial failures
	if result.Failed > 0 && result.Success > 0 {
		w.WriteHeader(http.StatusMultiStatus)
		Success(w, response, "Batch delete completed with some failures")
		return
	}

	// Return 400 if all failed
	if result.Failed > 0 && result.Success == 0 {
		w.WriteHeader(http.StatusBadRequest)
		Success(w, response, "All domains failed to delete")
		return
	}

	// Return 200 if all succeeded
	Success(w, response, "All domains deleted successfully")
}

func HandleBatchUpdateDomains(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Updates []struct {
			ID     int64                         `json:"id"`
			Update *models.DomainUpdateRequest  `json:"update"`
		} `json:"updates"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.Updates) == 0 {
		Error(w, http.StatusBadRequest, "No domain updates provided")
		return
	}

	// Convert to map for service layer
	updates := make(map[int64]*models.DomainUpdateRequest)
	for _, update := range req.Updates {
		updates[update.ID] = update.Update
	}

	result := domainService.BulkUpdateDomains(updates)

	response := map[string]interface{}{
		"total":           result.Total,
		"success":         result.Success,
		"failed":          result.Failed,
		"errors":          result.Errors,
		"success_domains": result.SuccessDomains,
	}

	// Return 207 Multi-Status if there were partial failures
	if result.Failed > 0 && result.Success > 0 {
		w.WriteHeader(http.StatusMultiStatus)
		Success(w, response, "Batch update completed with some failures")
		return
	}

	// Return 400 if all failed
	if result.Failed > 0 && result.Success == 0 {
		w.WriteHeader(http.StatusBadRequest)
		Success(w, response, "All domains failed to update")
		return
	}

	// Return 200 if all succeeded
	Success(w, response, "All domains updated successfully")
}

// Settings handlers

func HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := settingsService.GetSettings()
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to get settings")
		return
	}

	Success(w, settings, "")
}

func HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req models.SettingsUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := settingsService.UpdateSettings(&req); err != nil {
		Error(w, http.StatusInternalServerError, "Failed to update settings")
		return
	}

	Success(w, nil, "Settings updated successfully")
}

// Certificate handlers

func HandleListCertificates(w http.ResponseWriter, r *http.Request) {
	// 解析查詢參數
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// 獲取憑證列表
	certs, total, err := certificateService.ListCertificates(limit, offset)
	if err != nil {
		log.Printf("Failed to list certificates: %v", err)
		Error(w, http.StatusInternalServerError, "Failed to list certificates")
		return
	}

	response := map[string]interface{}{
		"certificates": certs,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
	}

	Success(w, response, "Certificates retrieved successfully")
}

func HandleUploadCertificate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DomainName     string `json:"domain_name"`
		CertificatePEM string `json:"certificate_pem"`
		PrivateKeyPEM  string `json:"private_key_pem"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 驗證必填欄位
	if req.DomainName == "" {
		Error(w, http.StatusBadRequest, "domain_name is required")
		return
	}
	if req.CertificatePEM == "" {
		Error(w, http.StatusBadRequest, "certificate_pem is required")
		return
	}
	if req.PrivateKeyPEM == "" {
		Error(w, http.StatusBadRequest, "private_key_pem is required")
		return
	}

	// 上傳憑證
	cert, err := certificateService.UploadCertificate(req.DomainName, req.CertificatePEM, req.PrivateKeyPEM)
	if err != nil {
		log.Printf("Failed to upload certificate: %v", err)
		Error(w, http.StatusBadRequest, fmt.Sprintf("Failed to upload certificate: %v", err))
		return
	}

	Success(w, cert, "Certificate uploaded successfully")
}

func HandleGetCertificate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid certificate ID")
		return
	}

	cert, err := certificateService.GetCertificateByID(id)
	if err != nil {
		log.Printf("Failed to get certificate: %v", err)
		Error(w, http.StatusNotFound, "Certificate not found")
		return
	}

	Success(w, cert, "Certificate retrieved successfully")
}

func HandleDeleteCertificate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid certificate ID")
		return
	}

	if err := certificateService.DeleteCertificate(id); err != nil {
		log.Printf("Failed to delete certificate: %v", err)
		Error(w, http.StatusInternalServerError, "Failed to delete certificate")
		return
	}

	Success(w, nil, "Certificate deleted successfully")
}

func HandleRenewCertificate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid certificate ID")
		return
	}

	// Renew the certificate
	cert, err := certificateService.RenewCertificate(id)
	if err != nil {
		log.Printf("Failed to renew certificate: %v", err)
		Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to renew certificate: %v", err))
		return
	}

	Success(w, cert, "Certificate renewed successfully")
}

func HandleGetExpiringCertificates(w http.ResponseWriter, r *http.Request) {
	// 解析查詢參數：幾天內到期，預設 30 天
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	// 獲取即將到期的憑證
	certs, err := certificateService.GetExpiringCertificates(days)
	if err != nil {
		log.Printf("Failed to get expiring certificates: %v", err)
		Error(w, http.StatusInternalServerError, "Failed to get expiring certificates")
		return
	}

	response := map[string]interface{}{
		"certificates": certs,
		"days":         days,
		"count":        len(certs),
	}

	Success(w, response, fmt.Sprintf("Found %d certificates expiring within %d days", len(certs), days))
}

// Service handlers

func HandleListServices(w http.ResponseWriter, r *http.Request) {
	serviceMgr := k8s.NewServiceManager()

	// 檢查是否要列出所有命名空間
	allNamespaces := r.URL.Query().Get("all_namespaces") == "true"

	if allNamespaces {
		// 列出所有命名空間
		namespaces, err := serviceMgr.ListNamespaces()
		if err != nil {
			log.Printf("Failed to list namespaces: %v", err)
			Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list namespaces: %v", err))
			return
		}

		response := map[string]interface{}{
			"namespaces": namespaces,
		}
		Success(w, response, "Namespaces retrieved successfully")
		return
	}

	// 解析查詢參數
	namespace := r.URL.Query().Get("namespace")

	// 如果沒有指定命名空間，列出所有命名空間的服務
	if namespace == "" {
		services, err := serviceMgr.ListAllServices()
		if err != nil {
			log.Printf("Failed to list all services: %v", err)
			Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list all services: %v", err))
			return
		}

		response := map[string]interface{}{
			"services": services,
			"count":    len(services),
		}

		Success(w, response, "All services retrieved successfully")
		return
	}

	// 使用 ServiceManager 列出特定命名空間的服務
	services, err := serviceMgr.ListServices(namespace)
	if err != nil {
		log.Printf("Failed to list services: %v", err)
		Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list services: %v", err))
		return
	}

	response := map[string]interface{}{
		"services":  services,
		"namespace": namespace,
		"count":     len(services),
	}

	Success(w, response, "Services retrieved successfully")
}

func HandleGetService(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")

	if namespace == "" || name == "" {
		Error(w, http.StatusBadRequest, "namespace and name are required")
		return
	}

	// 使用 ServiceManager 取得服務
	serviceMgr := k8s.NewServiceManager()
	service, err := serviceMgr.GetService(namespace, name)
	if err != nil {
		log.Printf("Failed to get service: %v", err)
		Error(w, http.StatusNotFound, fmt.Sprintf("Service not found: %v", err))
		return
	}

	Success(w, service, "Service retrieved successfully")
}

func HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	diagnosticSvc := services.NewDiagnosticService()

	health, err := diagnosticSvc.GetHealthStatus()
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to get health status")
		return
	}

	// Get system metrics
	metrics, _ := diagnosticSvc.GetSystemMetrics()

	response := map[string]interface{}{
		"health":  health,
		"metrics": metrics,
	}

	if health.Status == "unhealthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
		Success(w, response, "System is unhealthy")
	} else if health.Status == "degraded" {
		Success(w, response, "System is degraded")
	} else {
		Success(w, response, "All systems operational")
	}
}

func HandleGetLogs(w http.ResponseWriter, r *http.Request) {
	diagnosticSvc := services.NewDiagnosticService()

	// Parse query parameters
	level := r.URL.Query().Get("level")
	category := r.URL.Query().Get("category")

	limit := 50
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	filter := models.LogFilter{
		Level:  level,
		Category: category,
		Limit:  limit,
		Offset: offset,
	}

	// Parse time filters if provided
	if startTimeStr := r.URL.Query().Get("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filter.StartTime = t
		}
	}
	if endTimeStr := r.URL.Query().Get("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filter.EndTime = t
		}
	}

	logs, total, err := diagnosticSvc.GetLogs(filter)
	if err != nil {
		log.Printf("Failed to get logs: %v", err)
		Error(w, http.StatusInternalServerError, "Failed to retrieve logs")
		return
	}

	response := map[string]interface{}{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	Success(w, response, "Logs retrieved successfully")
}

func HandleListAPIKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := apiKeyService.ListAllKeys()
	if err != nil {
		Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list API keys: %v", err))
		return
	}
	Success(w, keys, "API keys retrieved successfully")
}

func HandleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req models.APIKeyCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Use admin ID 1 for now (single admin system)
	resp, err := apiKeyService.GenerateKey(&req, 1)
	if err != nil {
		if err == models.ErrInvalidInput {
			Error(w, http.StatusBadRequest, "Invalid input: key_name is required, permissions must be read/write/delete")
			return
		}
		Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create API key: %v", err))
		return
	}
	Success(w, resp, "API key created successfully. Save the raw_key now — it won't be shown again.")
}

func HandleDeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid API key ID")
		return
	}

	if err := apiKeyService.RevokeKey(id); err != nil {
		if err == models.ErrAPIKeyNotFound {
			Error(w, http.StatusNotFound, "API key not found")
			return
		}
		Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete API key: %v", err))
		return
	}
	Success(w, nil, "API key deleted successfully")
}

func HandleMCP(w http.ResponseWriter, r *http.Request) {
	if mcpServer == nil {
		Error(w, http.StatusInternalServerError, "MCP server not initialized")
		return
	}

	var body json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	response := mcpServer.HandleMessage(body)
	if response == nil {
		// Notification, no response needed
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func HandleCreateBackup(w http.ResponseWriter, r *http.Request) {
	info, err := backupService.CreateBackup()
	if err != nil {
		Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create backup: %v", err))
		return
	}
	Success(w, info, "Backup created successfully")
}

func HandleListBackups(w http.ResponseWriter, r *http.Request) {
	backups, err := backupService.ListBackups()
	if err != nil {
		Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list backups: %v", err))
		return
	}
	Success(w, backups, "Backups retrieved successfully")
}

func HandleDownloadBackup(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	path, err := backupService.GetBackupPath(filename)
	if err != nil {
		Error(w, http.StatusNotFound, fmt.Sprintf("Backup not found: %v", err))
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, path)
}

func HandleDeleteBackup(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if err := backupService.DeleteBackup(filename); err != nil {
		Error(w, http.StatusNotFound, fmt.Sprintf("Failed to delete backup: %v", err))
		return
	}
	Success(w, nil, "Backup deleted successfully")
}

func HandleSetCloudflareToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		APIToken string `json:"api_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.APIToken == "" {
		Error(w, http.StatusBadRequest, "api_token is required")
		return
	}

	if err := cloudflareService.SaveToken(req.APIToken); err != nil {
		Error(w, http.StatusBadRequest, fmt.Sprintf("Failed to save Cloudflare token: %v", err))
		return
	}
	Success(w, map[string]string{"status": "active"}, "Token validated and saved")
}

func HandleGetCloudflareStatus(w http.ResponseWriter, r *http.Request) {
	status, err := cloudflareService.GetStatus()
	if err != nil {
		Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get Cloudflare status: %v", err))
		return
	}
	Success(w, status, "")
}

func HandleDeleteCloudflareToken(w http.ResponseWriter, r *http.Request) {
	if err := cloudflareService.RemoveToken(); err != nil {
		Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to remove Cloudflare token: %v", err))
		return
	}
	Success(w, nil, "Cloudflare token removed")
}
