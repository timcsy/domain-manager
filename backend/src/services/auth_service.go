package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/domain-manager/backend/src/models"
	"github.com/domain-manager/backend/src/repositories"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication operations
type AuthService struct {
	adminRepo *repositories.AdminAccountRepository
	sessions  map[string]*Session // In-memory session storage (use Redis in production)
}

// Session represents a user session
type Session struct {
	Token     string
	UserID    int64
	Username  string
	ExpiresAt time.Time
}

// NewAuthService creates a new auth service
func NewAuthService(adminRepo *repositories.AdminAccountRepository) *AuthService {
	return &AuthService{
		adminRepo: adminRepo,
		sessions:  make(map[string]*Session),
	}
}

// Login authenticates a user and creates a session
func (s *AuthService) Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	// Get user by username
	account, err := s.adminRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, models.ErrInvalidCredentials
	}

	// Validate password
	if err := s.adminRepo.ValidatePassword(account, req.Password); err != nil {
		return nil, models.ErrInvalidCredentials
	}

	// Update last login
	if err := s.adminRepo.UpdateLastLogin(account.ID); err != nil {
		return nil, fmt.Errorf("failed to update last login: %w", err)
	}

	// Generate session token
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Create session
	expiresAt := time.Now().Add(24 * time.Hour)
	session := &Session{
		Token:     token,
		UserID:    account.ID,
		Username:  account.Username,
		ExpiresAt: expiresAt,
	}
	s.sessions[token] = session

	return &models.LoginResponse{
		Token:     token,
		User:      *account,
		ExpiresAt: expiresAt,
	}, nil
}

// Logout invalidates a session
func (s *AuthService) Logout(token string) error {
	delete(s.sessions, token)
	return nil
}

// ValidateToken validates a session token
func (s *AuthService) ValidateToken(token string) (*Session, error) {
	session, exists := s.sessions[token]
	if !exists {
		return nil, models.ErrUnauthorized
	}

	// Check expiration
	if time.Now().After(session.ExpiresAt) {
		delete(s.sessions, token)
		return nil, models.ErrUnauthorized
	}

	return session, nil
}

// ChangePassword validates old password and updates to new password
func (s *AuthService) ChangePassword(username, oldPassword, newPassword string) error {
	if len(newPassword) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}

	account, err := s.adminRepo.GetByUsername(username)
	if err != nil {
		return err
	}

	if err := s.adminRepo.ValidatePassword(account, oldPassword); err != nil {
		return fmt.Errorf("old password is incorrect")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.adminRepo.UpdatePassword(account.ID, string(hashedPassword)); err != nil {
		return err
	}

	// Clear all sessions to force re-login
	for k := range s.sessions {
		delete(s.sessions, k)
	}

	return nil
}

// UpdateEmail updates admin email
func (s *AuthService) UpdateEmail(username, email string) error {
	if email == "" || !strings.Contains(email, "@") {
		return fmt.Errorf("invalid email format")
	}

	account, err := s.adminRepo.GetByUsername(username)
	if err != nil {
		return err
	}

	return s.adminRepo.UpdateEmail(account.ID, email)
}

// GetProfile returns admin profile (username + email)
func (s *AuthService) GetProfile(username string) (map[string]string, error) {
	account, err := s.adminRepo.GetByUsername(username)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"username": account.Username,
		"email":    account.Email,
	}, nil
}

// generateToken generates a random session token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
