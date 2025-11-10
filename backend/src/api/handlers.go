package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/domain-manager/backend/src/db"
	"github.com/domain-manager/backend/src/k8s"
	"github.com/domain-manager/backend/src/models"
	"github.com/domain-manager/backend/src/repositories"
	"github.com/domain-manager/backend/src/services"
	"github.com/go-chi/chi/v5"
)

var (
	authService        *services.AuthService
	domainService      *services.DomainService
	certificateService *services.CertificateService
	settingsService    *services.SettingsService
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
		authService.Logout(token)
	}
	Success(w, nil, "Logout successful")
}

// Domain handlers

func HandleListDomains(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

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

// Service handlers

func HandleListServices(w http.ResponseWriter, r *http.Request) {
	// 解析查詢參數
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = "default"
	}

	// 使用 ServiceManager 列出服務
	serviceMgr := k8s.NewServiceManager()
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
	Success(w, []interface{}{}, "API keys feature coming soon")
}

func HandleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "Create API key not yet implemented")
}

func HandleDeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "Delete API key not yet implemented")
}

func HandleMCP(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "MCP endpoint not yet implemented")
}
