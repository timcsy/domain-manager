package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/domain-manager/backend/src/models"
)

// APIKeyRepository handles API key data operations
type APIKeyRepository struct {
	db *sql.DB
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *sql.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// Create creates a new API key
func (r *APIKeyRepository) Create(key *models.APIKey) error {
	query := `
		INSERT INTO api_keys (key_value, key_name, admin_id, permissions, enabled, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	result, err := r.db.Exec(query,
		key.KeyHash,
		key.KeyName,
		key.AdminID,
		key.Permissions,
		true,
		key.ExpiresAt,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create API key: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	key.ID = id
	key.CreatedAt = now
	key.Enabled = true

	return nil
}

// GetByKeyHash retrieves an API key by its hash
func (r *APIKeyRepository) GetByKeyHash(keyHash string) (*models.APIKey, error) {
	query := `SELECT id, key_value, key_name, admin_id, permissions, enabled, last_used_at, expires_at, created_at
		FROM api_keys WHERE key_value = ?`
	key := &models.APIKey{}
	err := r.db.QueryRow(query, keyHash).Scan(
		&key.ID,
		&key.KeyHash,
		&key.KeyName,
		&key.AdminID,
		&key.Permissions,
		&key.Enabled,
		&key.LastUsedAt,
		&key.ExpiresAt,
		&key.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, models.ErrAPIKeyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}
	return key, nil
}

// List retrieves all API keys for an admin
func (r *APIKeyRepository) List(adminID int64) ([]*models.APIKey, error) {
	query := `SELECT id, key_value, key_name, admin_id, permissions, enabled, last_used_at, expires_at, created_at
		FROM api_keys WHERE admin_id = ? ORDER BY created_at DESC`
	rows, err := r.db.Query(query, adminID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}
	defer rows.Close()

	keys := []*models.APIKey{}
	for rows.Next() {
		key := &models.APIKey{}
		err := rows.Scan(
			&key.ID,
			&key.KeyHash,
			&key.KeyName,
			&key.AdminID,
			&key.Permissions,
			&key.Enabled,
			&key.LastUsedAt,
			&key.ExpiresAt,
			&key.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}
		keys = append(keys, key)
	}

	return keys, nil
}

// ListAll retrieves all API keys
func (r *APIKeyRepository) ListAll() ([]*models.APIKey, error) {
	query := `SELECT id, key_value, key_name, admin_id, permissions, enabled, last_used_at, expires_at, created_at
		FROM api_keys ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}
	defer rows.Close()

	keys := []*models.APIKey{}
	for rows.Next() {
		key := &models.APIKey{}
		err := rows.Scan(
			&key.ID,
			&key.KeyHash,
			&key.KeyName,
			&key.AdminID,
			&key.Permissions,
			&key.Enabled,
			&key.LastUsedAt,
			&key.ExpiresAt,
			&key.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}
		keys = append(keys, key)
	}

	return keys, nil
}

// Delete deletes an API key
func (r *APIKeyRepository) Delete(id int64) error {
	query := `DELETE FROM api_keys WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return models.ErrAPIKeyNotFound
	}
	return nil
}

// UpdateLastUsed updates the last_used_at timestamp
func (r *APIKeyRepository) UpdateLastUsed(id int64) error {
	query := `UPDATE api_keys SET last_used_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update last used: %w", err)
	}
	return nil
}
