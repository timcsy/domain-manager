package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/domain-manager/backend/src/models"
)

// DiagnosticLogRepository handles diagnostic log data operations
type DiagnosticLogRepository struct {
	db *sql.DB
}

// NewDiagnosticLogRepository creates a new diagnostic log repository
func NewDiagnosticLogRepository(db *sql.DB) *DiagnosticLogRepository {
	return &DiagnosticLogRepository{db: db}
}

// Create inserts a new diagnostic log
func (r *DiagnosticLogRepository) Create(log *models.DiagnosticLog) error {
	query := `
		INSERT INTO diagnostic_logs (
			domain_id, domain_name, check_type, status,
			message, details, resolved, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	log.CreatedAt = time.Now()
	log.Resolved = false

	result, err := r.db.Exec(
		query,
		log.DomainID,
		log.DomainName,
		log.CheckType,
		log.Status,
		log.Message,
		log.Details,
		log.Resolved,
		log.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create diagnostic log: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	log.ID = id
	return nil
}

// List retrieves diagnostic logs with optional filters
func (r *DiagnosticLogRepository) List(filter models.DiagnosticLogFilter) ([]*models.DiagnosticLog, error) {
	query := `
		SELECT
			id, domain_id, domain_name, check_type, status,
			message, details, resolved, resolved_at, created_at
		FROM diagnostic_logs
		WHERE 1=1
	`
	args := []interface{}{}

	// Apply filters
	if filter.DomainID != nil {
		query += " AND domain_id = ?"
		args = append(args, *filter.DomainID)
	}

	if filter.CheckType != "" {
		query += " AND check_type = ?"
		args = append(args, filter.CheckType)
	}

	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	if filter.Resolved != nil {
		query += " AND resolved = ?"
		args = append(args, *filter.Resolved)
	}

	// Order by most recent first
	query += " ORDER BY created_at DESC"

	// Pagination
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list diagnostic logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.DiagnosticLog
	for rows.Next() {
		log := &models.DiagnosticLog{}
		err := rows.Scan(
			&log.ID,
			&log.DomainID,
			&log.DomainName,
			&log.CheckType,
			&log.Status,
			&log.Message,
			&log.Details,
			&log.Resolved,
			&log.ResolvedAt,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan diagnostic log: %w", err)
		}
		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating diagnostic logs: %w", err)
	}

	return logs, nil
}

// GetByID retrieves a diagnostic log by ID
func (r *DiagnosticLogRepository) GetByID(id int64) (*models.DiagnosticLog, error) {
	query := `
		SELECT
			id, domain_id, domain_name, check_type, status,
			message, details, resolved, resolved_at, created_at
		FROM diagnostic_logs
		WHERE id = ?
	`

	log := &models.DiagnosticLog{}
	err := r.db.QueryRow(query, id).Scan(
		&log.ID,
		&log.DomainID,
		&log.DomainName,
		&log.CheckType,
		&log.Status,
		&log.Message,
		&log.Details,
		&log.Resolved,
		&log.ResolvedAt,
		&log.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.ErrDiagnosticLogNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get diagnostic log: %w", err)
	}

	return log, nil
}

// MarkResolved marks a diagnostic log as resolved
func (r *DiagnosticLogRepository) MarkResolved(id int64) error {
	query := `
		UPDATE diagnostic_logs
		SET resolved = 1, resolved_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := r.db.Exec(query, now, id)
	if err != nil {
		return fmt.Errorf("failed to mark diagnostic log as resolved: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrDiagnosticLogNotFound
	}

	return nil
}

// MarkMultipleResolved marks multiple diagnostic logs as resolved
func (r *DiagnosticLogRepository) MarkMultipleResolved(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	// Build placeholders for SQL IN clause
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids)+1)
	args[0] = time.Now()

	for i, id := range ids {
		placeholders[i] = "?"
		args[i+1] = id
	}

	query := fmt.Sprintf(`
		UPDATE diagnostic_logs
		SET resolved = 1, resolved_at = ?
		WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to mark diagnostic logs as resolved: %w", err)
	}

	return nil
}

// Delete removes a diagnostic log by ID
func (r *DiagnosticLogRepository) Delete(id int64) error {
	query := `DELETE FROM diagnostic_logs WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete diagnostic log: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrDiagnosticLogNotFound
	}

	return nil
}

// Count counts diagnostic logs matching the filter
func (r *DiagnosticLogRepository) Count(filter models.DiagnosticLogFilter) (int, error) {
	query := `SELECT COUNT(*) FROM diagnostic_logs WHERE 1=1`
	args := []interface{}{}

	if filter.DomainID != nil {
		query += " AND domain_id = ?"
		args = append(args, *filter.DomainID)
	}

	if filter.CheckType != "" {
		query += " AND check_type = ?"
		args = append(args, filter.CheckType)
	}

	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	if filter.Resolved != nil {
		query += " AND resolved = ?"
		args = append(args, *filter.Resolved)
	}

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count diagnostic logs: %w", err)
	}

	return count, nil
}
