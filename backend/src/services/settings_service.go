package services

import (
	"fmt"

	"github.com/domain-manager/backend/src/models"
	"github.com/domain-manager/backend/src/repositories"
)

// SettingsService handles system settings operations
type SettingsService struct {
	settingsRepo *repositories.SettingsRepository
}

// NewSettingsService creates a new settings service
func NewSettingsService(settingsRepo *repositories.SettingsRepository) *SettingsService {
	return &SettingsService{
		settingsRepo: settingsRepo,
	}
}

// GetSettings retrieves all system settings
func (s *SettingsService) GetSettings() (map[string]*models.SystemSetting, error) {
	settings, err := s.settingsRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}
	return settings, nil
}

// UpdateSettings updates system settings
func (s *SettingsService) UpdateSettings(req *models.SettingsUpdateRequest) error {
	for key, value := range req.Settings {
		if err := s.settingsRepo.Set(key, value); err != nil {
			return fmt.Errorf("failed to update setting %s: %w", key, err)
		}
	}
	return nil
}

// GetSetting retrieves a single setting
func (s *SettingsService) GetSetting(key string) (*models.SystemSetting, error) {
	setting, err := s.settingsRepo.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get setting: %w", err)
	}
	return setting, nil
}
