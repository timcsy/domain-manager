package models

import (
	"time"
)

// APIKey represents an API key for programmatic access
type APIKey struct {
	ID        int64      `json:"id" db:"id"`
	KeyHash   string     `json:"-" db:"key_value"`
	KeyName   string     `json:"key_name" db:"key_name"`
	AdminID   int64      `json:"admin_id" db:"admin_id"`
	Permissions string   `json:"permissions" db:"permissions"`
	Enabled   bool       `json:"enabled" db:"enabled"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// APIKeyCreateRequest represents the request to create an API key
type APIKeyCreateRequest struct {
	KeyName     string   `json:"key_name"`
	Permissions []string `json:"permissions"`
	ExpiresAt   *string  `json:"expires_at,omitempty"`
}

// APIKeyResponse is returned after creating a key (includes the raw key, shown only once)
type APIKeyResponse struct {
	APIKey
	RawKey string `json:"raw_key,omitempty"`
}

// Validate validates the API key create request
func (r *APIKeyCreateRequest) Validate() error {
	if r.KeyName == "" {
		return ErrInvalidInput
	}
	if len(r.Permissions) == 0 {
		r.Permissions = []string{"read"}
	}
	for _, p := range r.Permissions {
		if p != "read" && p != "write" && p != "delete" {
			return ErrInvalidInput
		}
	}
	return nil
}
