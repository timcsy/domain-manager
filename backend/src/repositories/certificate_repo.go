package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/domain-manager/backend/src/models"
)

// CertificateRepository 處理憑證資料操作
type CertificateRepository struct {
	db *sql.DB
}

// NewCertificateRepository 建立新的憑證 repository
func NewCertificateRepository(db *sql.DB) *CertificateRepository {
	return &CertificateRepository{db: db}
}

// Create 建立新憑證
func (r *CertificateRepository) Create(cert *models.Certificate) error {
	query := `
		INSERT INTO certificates (
			domain_name, source, certificate_pem, private_key_pem, issuer,
			valid_from, valid_until, status, k8s_secret_name, k8s_secret_namespace,
			auto_renew, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	result, err := r.db.Exec(query,
		cert.DomainName,
		cert.Source,
		cert.CertificatePEM,
		cert.PrivateKeyPEM,
		cert.Issuer,
		cert.ValidFrom,
		cert.ValidUntil,
		cert.Status,
		cert.K8sSecretName,
		cert.K8sSecretNamespace,
		cert.AutoRenew,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	cert.ID = int(id)
	cert.CreatedAt = now
	cert.UpdatedAt = now

	return nil
}

// GetByID 根據 ID 取得憑證
func (r *CertificateRepository) GetByID(id int) (*models.Certificate, error) {
	query := `
		SELECT id, domain_name, source, certificate_pem, private_key_pem, issuer,
		       valid_from, valid_until, status, k8s_secret_name, k8s_secret_namespace,
		       auto_renew, last_renewal_attempt, renewal_error, created_at, updated_at
		FROM certificates
		WHERE id = ?
	`
	cert := &models.Certificate{}
	err := r.db.QueryRow(query, id).Scan(
		&cert.ID,
		&cert.DomainName,
		&cert.Source,
		&cert.CertificatePEM,
		&cert.PrivateKeyPEM,
		&cert.Issuer,
		&cert.ValidFrom,
		&cert.ValidUntil,
		&cert.Status,
		&cert.K8sSecretName,
		&cert.K8sSecretNamespace,
		&cert.AutoRenew,
		&cert.LastRenewalAttempt,
		&cert.RenewalError,
		&cert.CreatedAt,
		&cert.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("certificate not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	}

	return cert, nil
}

// GetByDomain 根據域名取得憑證
func (r *CertificateRepository) GetByDomain(domainName string) (*models.Certificate, error) {
	query := `
		SELECT id, domain_name, source, certificate_pem, private_key_pem, issuer,
		       valid_from, valid_until, status, k8s_secret_name, k8s_secret_namespace,
		       auto_renew, last_renewal_attempt, renewal_error, created_at, updated_at
		FROM certificates
		WHERE domain_name = ?
		ORDER BY created_at DESC
		LIMIT 1
	`
	cert := &models.Certificate{}
	err := r.db.QueryRow(query, domainName).Scan(
		&cert.ID,
		&cert.DomainName,
		&cert.Source,
		&cert.CertificatePEM,
		&cert.PrivateKeyPEM,
		&cert.Issuer,
		&cert.ValidFrom,
		&cert.ValidUntil,
		&cert.Status,
		&cert.K8sSecretName,
		&cert.K8sSecretNamespace,
		&cert.AutoRenew,
		&cert.LastRenewalAttempt,
		&cert.RenewalError,
		&cert.CreatedAt,
		&cert.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("certificate not found for domain: %s", domainName)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	}

	return cert, nil
}

// List 列出所有憑證，支援分頁
func (r *CertificateRepository) List(limit, offset int) ([]*models.Certificate, error) {
	query := `
		SELECT id, domain_name, source, certificate_pem, private_key_pem, issuer,
		       valid_from, valid_until, status, k8s_secret_name, k8s_secret_namespace,
		       auto_renew, last_renewal_attempt, renewal_error, created_at, updated_at
		FROM certificates
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list certificates: %w", err)
	}
	defer rows.Close()

	var certs []*models.Certificate
	for rows.Next() {
		cert := &models.Certificate{}
		err := rows.Scan(
			&cert.ID,
			&cert.DomainName,
			&cert.Source,
			&cert.CertificatePEM,
			&cert.PrivateKeyPEM,
			&cert.Issuer,
			&cert.ValidFrom,
			&cert.ValidUntil,
			&cert.Status,
			&cert.K8sSecretName,
			&cert.K8sSecretNamespace,
			&cert.AutoRenew,
			&cert.LastRenewalAttempt,
			&cert.RenewalError,
			&cert.CreatedAt,
			&cert.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan certificate: %w", err)
		}
		certs = append(certs, cert)
	}

	return certs, nil
}

// Update 更新憑證
func (r *CertificateRepository) Update(cert *models.Certificate) error {
	query := `
		UPDATE certificates
		SET certificate_pem = ?, private_key_pem = ?, issuer = ?,
		    valid_from = ?, valid_until = ?, status = ?,
		    k8s_secret_name = ?, k8s_secret_namespace = ?,
		    auto_renew = ?, last_renewal_attempt = ?, renewal_error = ?,
		    updated_at = ?
		WHERE id = ?
	`
	now := time.Now()
	_, err := r.db.Exec(query,
		cert.CertificatePEM,
		cert.PrivateKeyPEM,
		cert.Issuer,
		cert.ValidFrom,
		cert.ValidUntil,
		cert.Status,
		cert.K8sSecretName,
		cert.K8sSecretNamespace,
		cert.AutoRenew,
		cert.LastRenewalAttempt,
		cert.RenewalError,
		now,
		cert.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update certificate: %w", err)
	}

	cert.UpdatedAt = now
	return nil
}

// Delete 刪除憑證
func (r *CertificateRepository) Delete(id int) error {
	query := `DELETE FROM certificates WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete certificate: %w", err)
	}
	return nil
}

// GetExpiring 取得即將到期的憑證（30天內）
func (r *CertificateRepository) GetExpiring(days int) ([]*models.Certificate, error) {
	query := `
		SELECT id, domain_name, source, certificate_pem, private_key_pem, issuer,
		       valid_from, valid_until, status, k8s_secret_name, k8s_secret_namespace,
		       auto_renew, last_renewal_attempt, renewal_error, created_at, updated_at
		FROM certificates
		WHERE valid_until <= datetime('now', '+' || ? || ' days')
		  AND status != 'expired'
		  AND status != 'revoked'
		ORDER BY valid_until ASC
	`
	rows, err := r.db.Query(query, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get expiring certificates: %w", err)
	}
	defer rows.Close()

	var certs []*models.Certificate
	for rows.Next() {
		cert := &models.Certificate{}
		err := rows.Scan(
			&cert.ID,
			&cert.DomainName,
			&cert.Source,
			&cert.CertificatePEM,
			&cert.PrivateKeyPEM,
			&cert.Issuer,
			&cert.ValidFrom,
			&cert.ValidUntil,
			&cert.Status,
			&cert.K8sSecretName,
			&cert.K8sSecretNamespace,
			&cert.AutoRenew,
			&cert.LastRenewalAttempt,
			&cert.RenewalError,
			&cert.CreatedAt,
			&cert.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan certificate: %w", err)
		}
		certs = append(certs, cert)
	}

	return certs, nil
}

// Count 計算憑證總數
func (r *CertificateRepository) Count() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM certificates`
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count certificates: %w", err)
	}
	return count, nil
}

// UpdateRenewalStatus 更新續約狀態
func (r *CertificateRepository) UpdateRenewalStatus(id int, lastAttempt time.Time, renewalError *string) error {
	query := `
		UPDATE certificates
		SET last_renewal_attempt = ?, renewal_error = ?, updated_at = ?
		WHERE id = ?
	`
	now := time.Now()
	_, err := r.db.Exec(query, lastAttempt, renewalError, now, id)
	if err != nil {
		return fmt.Errorf("failed to update renewal status: %w", err)
	}
	return nil
}
