package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/domain-manager/backend/src/models"
)

// DomainRepository handles domain data operations
type DomainRepository struct {
	db *sql.DB
}

// NewDomainRepository creates a new domain repository
func NewDomainRepository(db *sql.DB) *DomainRepository {
	return &DomainRepository{db: db}
}

// Create creates a new domain
func (r *DomainRepository) Create(domain *models.Domain) error {
	query := `
		INSERT INTO domains (domain_name, target_service, target_namespace, target_port, ssl_mode, status, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	result, err := r.db.Exec(query,
		domain.DomainName,
		domain.TargetService,
		domain.TargetNamespace,
		domain.TargetPort,
		domain.SSLMode,
		"pending",
		true,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create domain: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	domain.ID = id
	domain.CreatedAt = now
	domain.UpdatedAt = now
	domain.Status = "pending"
	domain.Enabled = true

	return nil
}

// GetByID retrieves a domain by ID
func (r *DomainRepository) GetByID(id int64) (*models.Domain, error) {
	query := `SELECT * FROM domains WHERE id = ?`
	domain := &models.Domain{}
	err := r.db.QueryRow(query, id).Scan(
		&domain.ID,
		&domain.DomainName,
		&domain.TargetService,
		&domain.TargetNamespace,
		&domain.TargetPort,
		&domain.SSLMode,
		&domain.CertificateID,
		&domain.Status,
		&domain.Enabled,
		&domain.CreatedAt,
		&domain.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, models.ErrDomainNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}
	return domain, nil
}

// GetByName retrieves a domain by name
func (r *DomainRepository) GetByName(name string) (*models.Domain, error) {
	query := `SELECT * FROM domains WHERE domain_name = ?`
	domain := &models.Domain{}
	err := r.db.QueryRow(query, name).Scan(
		&domain.ID,
		&domain.DomainName,
		&domain.TargetService,
		&domain.TargetNamespace,
		&domain.TargetPort,
		&domain.SSLMode,
		&domain.CertificateID,
		&domain.Status,
		&domain.Enabled,
		&domain.CreatedAt,
		&domain.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, models.ErrDomainNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}
	return domain, nil
}

// List retrieves domains with filters
func (r *DomainRepository) List(filter models.DomainFilter) ([]*models.Domain, error) {
	query := `SELECT * FROM domains WHERE 1=1`
	args := []interface{}{}

	if filter.Status != "" {
		query += ` AND status = ?`
		args = append(args, filter.Status)
	}
	if filter.Enabled != nil {
		query += ` AND enabled = ?`
		args = append(args, *filter.Enabled)
	}
	if filter.ServiceName != "" {
		query += ` AND target_service = ?`
		args = append(args, filter.ServiceName)
	}
	if filter.Namespace != "" {
		query += ` AND target_namespace = ?`
		args = append(args, filter.Namespace)
	}

	query += ` ORDER BY created_at DESC`

	if filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += ` OFFSET ?`
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}
	defer rows.Close()

	domains := []*models.Domain{}
	for rows.Next() {
		domain := &models.Domain{}
		err := rows.Scan(
			&domain.ID,
			&domain.DomainName,
			&domain.TargetService,
			&domain.TargetNamespace,
			&domain.TargetPort,
			&domain.SSLMode,
			&domain.CertificateID,
			&domain.Status,
			&domain.Enabled,
			&domain.CreatedAt,
			&domain.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan domain: %w", err)
		}
		domains = append(domains, domain)
	}

	return domains, nil
}

// Update updates a domain
func (r *DomainRepository) Update(domain *models.Domain) error {
	query := `
		UPDATE domains
		SET target_service = ?, target_namespace = ?, target_port = ?,
		    ssl_mode = ?, certificate_id = ?, status = ?, enabled = ?, updated_at = ?
		WHERE id = ?
	`
	domain.UpdatedAt = time.Now()
	_, err := r.db.Exec(query,
		domain.TargetService,
		domain.TargetNamespace,
		domain.TargetPort,
		domain.SSLMode,
		domain.CertificateID,
		domain.Status,
		domain.Enabled,
		domain.UpdatedAt,
		domain.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update domain: %w", err)
	}
	return nil
}

// Delete hard deletes a domain
func (r *DomainRepository) Delete(id int64) error {
	query := `DELETE FROM domains WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete domain: %w", err)
	}
	return nil
}

// SoftDelete marks a domain as deleted
func (r *DomainRepository) SoftDelete(id int64) error {
	query := `UPDATE domains SET enabled = 0, status = 'deleted', updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to soft delete domain: %w", err)
	}
	return nil
}

// Count counts domains matching the filter
func (r *DomainRepository) Count(filter models.DomainFilter) (int, error) {
	query := `SELECT COUNT(*) FROM domains WHERE 1=1`
	args := []interface{}{}

	if filter.Status != "" {
		query += ` AND status = ?`
		args = append(args, filter.Status)
	}
	if filter.Enabled != nil {
		query += ` AND enabled = ?`
		args = append(args, *filter.Enabled)
	}

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count domains: %w", err)
	}
	return count, nil
}
