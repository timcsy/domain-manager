package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/domain-manager/backend/src/models"
	"golang.org/x/crypto/bcrypt"
)

// AdminAccountRepository handles admin account data operations
type AdminAccountRepository struct {
	db *sql.DB
}

// NewAdminAccountRepository creates a new admin account repository
func NewAdminAccountRepository(db *sql.DB) *AdminAccountRepository {
	return &AdminAccountRepository{db: db}
}

// Create creates a new admin account
func (r *AdminAccountRepository) Create(account *models.AdminAccount, password string) error {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		INSERT INTO admin_accounts (username, password_hash, email, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`
	now := time.Now()
	result, err := r.db.Exec(query, account.Username, string(hashedPassword), account.Email, now, now)
	if err != nil {
		return fmt.Errorf("failed to create admin account: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	account.ID = id
	account.CreatedAt = now
	account.UpdatedAt = now

	return nil
}

// GetByUsername retrieves an admin account by username
func (r *AdminAccountRepository) GetByUsername(username string) (*models.AdminAccount, error) {
	query := `SELECT id, username, password_hash, email, last_login_at, created_at, updated_at FROM admin_accounts WHERE username = ?`
	account := &models.AdminAccount{}
	err := r.db.QueryRow(query, username).Scan(
		&account.ID,
		&account.Username,
		&account.PasswordHash,
		&account.Email,
		&account.LastLoginAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, models.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get admin account: %w", err)
	}
	return account, nil
}

// ValidatePassword validates a password against the stored hash
func (r *AdminAccountRepository) ValidatePassword(account *models.AdminAccount, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password))
	if err != nil {
		return models.ErrInvalidCredentials
	}
	return nil
}

// GetByID retrieves an admin account by ID
func (r *AdminAccountRepository) GetByID(id int64) (*models.AdminAccount, error) {
	query := `SELECT id, username, password_hash, email, last_login_at, created_at, updated_at FROM admin_accounts WHERE id = ?`
	account := &models.AdminAccount{}
	err := r.db.QueryRow(query, id).Scan(
		&account.ID,
		&account.Username,
		&account.PasswordHash,
		&account.Email,
		&account.LastLoginAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, models.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get admin account: %w", err)
	}
	return account, nil
}

// UpdatePassword updates the password hash for an admin account
func (r *AdminAccountRepository) UpdatePassword(id int64, newPasswordHash string) error {
	query := `UPDATE admin_accounts SET password_hash = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, newPasswordHash, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

// UpdateEmail updates the email for an admin account
func (r *AdminAccountRepository) UpdateEmail(id int64, email string) error {
	query := `UPDATE admin_accounts SET email = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, email, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update email: %w", err)
	}
	return nil
}

// UpdateLastLogin updates the last login timestamp
func (r *AdminAccountRepository) UpdateLastLogin(id int64) error {
	query := `UPDATE admin_accounts SET last_login_at = ? WHERE id = ?`
	now := time.Now()
	_, err := r.db.Exec(query, now, id)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}
