package services

import (
	"testing"

	"github.com/domain-manager/backend/src/models"
)

// TestBulkOperationResult tests the BulkOperationResult struct
func TestBulkOperationResult(t *testing.T) {
	t.Run("Initialize bulk operation result", func(t *testing.T) {
		result := &BulkOperationResult{
			Total:          3,
			Success:        2,
			Failed:         1,
			Errors:         []BulkOperationError{},
			SuccessDomains: []*models.Domain{},
		}

		if result.Total != 3 {
			t.Errorf("Expected total 3, got %d", result.Total)
		}
		if result.Success != 2 {
			t.Errorf("Expected success 2, got %d", result.Success)
		}
		if result.Failed != 1 {
			t.Errorf("Expected failed 1, got %d", result.Failed)
		}
	})

	t.Run("Add error to bulk operation result", func(t *testing.T) {
		result := &BulkOperationResult{
			Total:          0,
			Success:        0,
			Failed:         0,
			Errors:         []BulkOperationError{},
			SuccessDomains: []*models.Domain{},
		}

		result.Errors = append(result.Errors, BulkOperationError{
			DomainName: "test.com",
			Error:      "test error",
		})

		if len(result.Errors) != 1 {
			t.Errorf("Expected 1 error, got %d", len(result.Errors))
		}
		if result.Errors[0].DomainName != "test.com" {
			t.Errorf("Expected domain name 'test.com', got '%s'", result.Errors[0].DomainName)
		}
	})

	t.Run("Add success domain to bulk operation result", func(t *testing.T) {
		result := &BulkOperationResult{
			Total:          0,
			Success:        0,
			Failed:         0,
			Errors:         []BulkOperationError{},
			SuccessDomains: []*models.Domain{},
		}

		domain := &models.Domain{
			ID:         1,
			DomainName: "test.com",
		}
		result.SuccessDomains = append(result.SuccessDomains, domain)

		if len(result.SuccessDomains) != 1 {
			t.Errorf("Expected 1 success domain, got %d", len(result.SuccessDomains))
		}
		if result.SuccessDomains[0].DomainName != "test.com" {
			t.Errorf("Expected domain name 'test.com', got '%s'", result.SuccessDomains[0].DomainName)
		}
	})
}

// TestBulkOperationError tests the BulkOperationError struct
func TestBulkOperationError(t *testing.T) {
	t.Run("Create bulk operation error", func(t *testing.T) {
		err := BulkOperationError{
			DomainName: "example.com",
			Error:      "validation failed",
		}

		if err.DomainName != "example.com" {
			t.Errorf("Expected domain name 'example.com', got '%s'", err.DomainName)
		}
		if err.Error != "validation failed" {
			t.Errorf("Expected error 'validation failed', got '%s'", err.Error)
		}
	})
}
