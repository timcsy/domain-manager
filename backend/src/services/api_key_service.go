package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/domain-manager/backend/src/models"
	"github.com/domain-manager/backend/src/repositories"
)

// APIKeyService handles API key business logic
type APIKeyService struct {
	repo *repositories.APIKeyRepository
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService(repo *repositories.APIKeyRepository) *APIKeyService {
	return &APIKeyService{repo: repo}
}

// GenerateKey creates a new API key and returns it (raw key is only available at creation time)
func (s *APIKeyService) GenerateKey(req *models.APIKeyCreateRequest, adminID int64) (*models.APIKeyResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Generate random key
	rawKey, err := generateRandomKey(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Hash the key for storage
	keyHash := hashKey(rawKey)

	// Serialize permissions
	permJSON, err := json.Marshal(req.Permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize permissions: %w", err)
	}

	// Parse expiration
	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("invalid expires_at format, use RFC3339: %w", err)
		}
		expiresAt = &t
	}

	key := &models.APIKey{
		KeyHash:     keyHash,
		KeyName:     req.KeyName,
		AdminID:     adminID,
		Permissions: string(permJSON),
		ExpiresAt:   expiresAt,
	}

	if err := s.repo.Create(key); err != nil {
		return nil, err
	}

	return &models.APIKeyResponse{
		APIKey: *key,
		RawKey: rawKey,
	}, nil
}

// ValidateKey validates an API key and returns the key record
func (s *APIKeyService) ValidateKey(rawKey string) (*models.APIKey, error) {
	keyHash := hashKey(rawKey)
	key, err := s.repo.GetByKeyHash(keyHash)
	if err != nil {
		return nil, err
	}

	if !key.Enabled {
		return nil, models.ErrAPIKeyDisabled
	}

	if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
		return nil, models.ErrAPIKeyExpired
	}

	// Update last used timestamp asynchronously
	go func() {
		_ = s.repo.UpdateLastUsed(key.ID)
	}()

	return key, nil
}

// ListKeys returns all API keys for an admin
func (s *APIKeyService) ListKeys(adminID int64) ([]*models.APIKey, error) {
	return s.repo.List(adminID)
}

// ListAllKeys returns all API keys
func (s *APIKeyService) ListAllKeys() ([]*models.APIKey, error) {
	return s.repo.ListAll()
}

// RevokeKey deletes an API key
func (s *APIKeyService) RevokeKey(id int64) error {
	return s.repo.Delete(id)
}

func generateRandomKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "dm_" + hex.EncodeToString(bytes), nil
}

func hashKey(rawKey string) string {
	hash := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(hash[:])
}
