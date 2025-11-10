package models

import (
	"time"
)

// SystemSetting represents a system configuration setting
type SystemSetting struct {
	Key         string    `json:"key" db:"key"`
	Value       string    `json:"value" db:"value"`
	Description string    `json:"description" db:"description"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// SettingsUpdateRequest represents a request to update settings
type SettingsUpdateRequest struct {
	Settings map[string]string `json:"settings" validate:"required"`
}
