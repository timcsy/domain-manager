package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/domain-manager/backend/src/k8s"
	"github.com/domain-manager/backend/src/models"
)

// CloudflareService handles Cloudflare DNS integration
type CloudflareService struct {
	settingsService   *SettingsService
	certManagerHelper *k8s.CertManagerHelper
	certManagerNS     string
}

// NewCloudflareService creates a new CloudflareService
func NewCloudflareService(settingsService *SettingsService) *CloudflareService {
	ns := "cert-manager"
	return &CloudflareService{
		settingsService:   settingsService,
		certManagerHelper: k8s.NewCertManagerHelper(),
		certManagerNS:     ns,
	}
}

// CloudflareStatus represents the current Cloudflare integration status
type CloudflareStatus struct {
	Enabled            bool `json:"enabled"`
	TokenSet           bool `json:"token_set"`
	TokenValid         bool `json:"token_valid"`
	ClusterIssuerReady bool `json:"cluster_issuer_ready"`
}

// ValidateToken validates a Cloudflare API token against the Cloudflare API
func (s *CloudflareService) ValidateToken(token string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("token is empty")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://api.cloudflare.com/client/v4/user/tokens/verify", nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to verify token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("Cloudflare token verify response: status=%d body=%s", resp.StatusCode, string(body))

	if resp.StatusCode == 200 {
		var result struct {
			Success bool `json:"success"`
			Result  struct {
				Status string `json:"status"`
			} `json:"result"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return false, fmt.Errorf("failed to parse response: %w", err)
		}
		return result.Success && result.Result.Status == "active", nil
	}

	return false, fmt.Errorf("Cloudflare API returned status %d", resp.StatusCode)
}

// SaveToken validates, saves the token, creates K8s Secret, and updates ClusterIssuer
func (s *CloudflareService) SaveToken(token string) error {
	// Validate first
	valid, err := s.ValidateToken(token)
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid Cloudflare API token")
	}

	// Save to system settings
	err = s.settingsService.UpdateSettings(&models.SettingsUpdateRequest{
		Settings: map[string]string{
			"cloudflare_api_token": token,
			"cloudflare_enabled":  "1",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to save token to settings: %w", err)
	}

	// Create K8s Secret
	if err := s.certManagerHelper.CreateOrUpdateCloudflareSecret(s.certManagerNS, token); err != nil {
		log.Printf("⚠️  Failed to create K8s Secret (may work later in cluster): %v", err)
	}

	// Update ClusterIssuer with DNS-01 solver
	if err := s.updateClusterIssuer(true); err != nil {
		log.Printf("⚠️  Failed to update ClusterIssuer: %v", err)
	}

	log.Printf("✅ Cloudflare API token saved and DNS-01 configured")
	return nil
}

// RemoveToken removes the Cloudflare token and reverts to HTTP-01 only
func (s *CloudflareService) RemoveToken() error {
	err := s.settingsService.UpdateSettings(&models.SettingsUpdateRequest{
		Settings: map[string]string{
			"cloudflare_api_token": "",
			"cloudflare_enabled":  "0",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to remove token from settings: %w", err)
	}

	// Delete K8s Secret
	if err := s.certManagerHelper.DeleteCloudflareSecret(s.certManagerNS); err != nil {
		log.Printf("⚠️  Failed to delete K8s Secret: %v", err)
	}

	// Revert ClusterIssuer to HTTP-01 only
	if err := s.updateClusterIssuer(false); err != nil {
		log.Printf("⚠️  Failed to revert ClusterIssuer: %v", err)
	}

	log.Printf("✅ Cloudflare token removed, reverted to HTTP-01")
	return nil
}

// GetStatus returns the current Cloudflare integration status
func (s *CloudflareService) GetStatus() (*CloudflareStatus, error) {
	status := &CloudflareStatus{}

	enabledSetting, err := s.settingsService.GetSetting("cloudflare_enabled")
	if err == nil && enabledSetting != nil {
		status.Enabled = enabledSetting.Value == "1"
	}

	tokenSetting, err := s.settingsService.GetSetting("cloudflare_api_token")
	if err == nil && tokenSetting != nil {
		status.TokenSet = tokenSetting.Value != ""

		if status.TokenSet {
			valid, _ := s.ValidateToken(tokenSetting.Value)
			status.TokenValid = valid
		}
	}

	ready, _ := s.certManagerHelper.GetClusterIssuerStatus("letsencrypt-prod")
	status.ClusterIssuerReady = ready

	return status, nil
}

func (s *CloudflareService) updateClusterIssuer(cloudflareEnabled bool) error {
	// Read settings for ClusterIssuer config
	email := ""
	acmeServer := "https://acme-v02.api.letsencrypt.org/directory"
	ingressClass := "nginx"

	if setting, err := s.settingsService.GetSetting("letsencrypt_email"); err == nil && setting != nil {
		email = setting.Value
	}
	if setting, err := s.settingsService.GetSetting("letsencrypt_server"); err == nil && setting != nil && setting.Value != "" {
		acmeServer = setting.Value
	}
	if setting, err := s.settingsService.GetSetting("default_ingress_class"); err == nil && setting != nil && setting.Value != "" {
		ingressClass = setting.Value
	}

	cfg := &k8s.ClusterIssuerConfig{
		Name:               "letsencrypt-prod",
		Email:              email,
		ACMEServer:         acmeServer,
		IngressClass:       ingressClass,
		CloudflareEnabled:  cloudflareEnabled,
		CloudflareSecretNS: s.certManagerNS,
	}

	return s.certManagerHelper.CreateOrUpdateClusterIssuer(cfg)
}
