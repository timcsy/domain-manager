package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/domain-manager/backend/src/models"
)

// SettingsRepository handles system settings operations
type SettingsRepository struct {
	db *sql.DB
}

// NewSettingsRepository creates a new settings repository
func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// Get retrieves a setting by key
func (r *SettingsRepository) Get(key string) (*models.SystemSetting, error) {
	query := `SELECT key, value, description, updated_at FROM system_settings WHERE key = ?`
	setting := &models.SystemSetting{}
	err := r.db.QueryRow(query, key).Scan(
		&setting.Key,
		&setting.Value,
		&setting.Description,
		&setting.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get setting: %w", err)
	}
	return setting, nil
}

// Set creates or updates a setting
func (r *SettingsRepository) Set(key, value string) error {
	query := `
		INSERT INTO system_settings (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = ?
	`
	now := time.Now()
	_, err := r.db.Exec(query, key, value, now, value, now)
	if err != nil {
		return fmt.Errorf("failed to set setting: %w", err)
	}
	return nil
}

// GetAll retrieves all settings
func (r *SettingsRepository) GetAll() (map[string]*models.SystemSetting, error) {
	query := `SELECT key, value, description, updated_at FROM system_settings`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all settings: %w", err)
	}
	defer rows.Close()

	settings := make(map[string]*models.SystemSetting)
	for rows.Next() {
		setting := &models.SystemSetting{}
		err := rows.Scan(&setting.Key, &setting.Value, &setting.Description, &setting.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan setting: %w", err)
		}
		settings[setting.Key] = setting
	}

	return settings, nil
}
