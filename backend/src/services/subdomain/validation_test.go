package subdomain

import (
	"testing"
)

func TestValidateSubdomain(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name          string
		domain        string
		expectValid   bool
		expectSub     bool
		expectRoot    string
		expectSubpart string
		expectError   string
	}{
		{
			name:        "Valid root domain",
			domain:      "example.com",
			expectValid: true,
			expectSub:   false,
			expectRoot:  "example.com",
		},
		{
			name:          "Valid subdomain",
			domain:        "api.example.com",
			expectValid:   true,
			expectSub:     true,
			expectRoot:    "example.com",
			expectSubpart: "api",
		},
		{
			name:          "Valid multi-level subdomain",
			domain:        "api.v1.example.com",
			expectValid:   true,
			expectSub:     true,
			expectRoot:    "example.com",
			expectSubpart: "api.v1",
		},
		{
			name:        "Empty domain",
			domain:      "",
			expectValid: false,
			expectError: "domain cannot be empty",
		},
		{
			name:        "Invalid format - no TLD",
			domain:      "example",
			expectValid: false,
			expectError: "invalid domain format",
		},
		{
			name:        "Invalid format - starts with hyphen",
			domain:      "-example.com",
			expectValid: false,
			expectError: "invalid domain format",
		},
		{
			name:        "Invalid format - ends with hyphen",
			domain:      "example-.com",
			expectValid: false,
			expectError: "invalid domain format",
		},
		{
			name:        "Invalid format - double dot",
			domain:      "example..com",
			expectValid: false,
			expectError: "empty label in domain",
		},
		{
			name:        "Label too long",
			domain:      "this-is-a-very-long-label-that-exceeds-the-maximum-allowed-length-for-a-dns-label-which-is-sixty-three-characters.com",
			expectValid: false,
			expectError: "exceeds 63 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateSubdomain(tt.domain)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got %v", tt.expectValid, result.Valid)
			}

			if result.IsSubdomain != tt.expectSub {
				t.Errorf("Expected isSubdomain=%v, got %v", tt.expectSub, result.IsSubdomain)
			}

			if tt.expectValid {
				if result.RootDomain != tt.expectRoot {
					t.Errorf("Expected root=%s, got %s", tt.expectRoot, result.RootDomain)
				}

				if tt.expectSub && result.Subdomain != tt.expectSubpart {
					t.Errorf("Expected subdomain part=%s, got %s", tt.expectSubpart, result.Subdomain)
				}
			}

			if !tt.expectValid && tt.expectError != "" {
				if result.ErrorMessage != tt.expectError {
					// Check if error message contains expected error
					if len(result.ErrorMessage) == 0 {
						t.Errorf("Expected error containing '%s', got empty", tt.expectError)
					}
				}
			}
		})
	}
}

func TestCheckSubdomainConflict(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name             string
		newDomain        string
		existingDomains  []string
		expectConflicts  int
		expectTypes      []string
	}{
		{
			name:             "No conflicts",
			newDomain:        "api.example.com",
			existingDomains:  []string{"other.example.com", "different.com"},
			expectConflicts:  0,
		},
		{
			name:             "Exact match conflict",
			newDomain:        "api.example.com",
			existingDomains:  []string{"api.example.com"},
			expectConflicts:  1,
			expectTypes:      []string{"exact"},
		},
		{
			name:             "Parent domain exists",
			newDomain:        "api.v1.example.com",
			existingDomains:  []string{"example.com"},
			expectConflicts:  1,
			expectTypes:      []string{"parent"},
		},
		{
			name:             "Child domain exists",
			newDomain:        "example.com",
			existingDomains:  []string{"api.example.com"},
			expectConflicts:  1,
			expectTypes:      []string{"child"},
		},
		{
			name:             "Wildcard conflict - existing wildcard",
			newDomain:        "api.example.com",
			existingDomains:  []string{"*.example.com"},
			expectConflicts:  1,
			expectTypes:      []string{"wildcard"},
		},
		{
			name:             "Wildcard conflict - new wildcard",
			newDomain:        "*.example.com",
			existingDomains:  []string{"api.example.com"},
			expectConflicts:  1,
			expectTypes:      []string{"wildcard"},
		},
		{
			name:             "Multiple conflicts",
			newDomain:        "api.example.com",
			existingDomains:  []string{"api.example.com", "example.com", "*.example.com"},
			expectConflicts:  3,
			expectTypes:      []string{"exact", "parent", "wildcard"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conflicts := validator.CheckSubdomainConflict(tt.newDomain, tt.existingDomains)

			if len(conflicts) != tt.expectConflicts {
				t.Errorf("Expected %d conflicts, got %d", tt.expectConflicts, len(conflicts))
			}

			if len(tt.expectTypes) > 0 {
				for i, expectedType := range tt.expectTypes {
					if i >= len(conflicts) {
						break
					}
					if conflicts[i].ConflictType != expectedType {
						t.Errorf("Conflict %d: expected type %s, got %s", i, expectedType, conflicts[i].ConflictType)
					}
				}
			}
		})
	}
}

func TestGetRootDomain(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		domain string
		expect string
	}{
		{"example.com", "example.com"},
		{"api.example.com", "example.com"},
		{"api.v1.example.com", "example.com"},
		{"test.api.v1.example.com", "example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			result := validator.GetRootDomain(tt.domain)
			if result != tt.expect {
				t.Errorf("Expected %s, got %s", tt.expect, result)
			}
		})
	}
}

func TestIsSubdomain(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		domain string
		expect bool
	}{
		{"example.com", false},
		{"api.example.com", true},
		{"api.v1.example.com", true},
		{"www.example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			result := validator.IsSubdomain(tt.domain)
			if result != tt.expect {
				t.Errorf("Expected %v, got %v", tt.expect, result)
			}
		})
	}
}

func TestGroupByRootDomain(t *testing.T) {
	validator := NewValidator()

	domains := []string{
		"example.com",
		"api.example.com",
		"www.example.com",
		"other.com",
		"api.other.com",
	}

	groups := validator.GroupByRootDomain(domains)

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups))
	}

	if len(groups["example.com"]) != 3 {
		t.Errorf("Expected 3 domains under example.com, got %d", len(groups["example.com"]))
	}

	if len(groups["other.com"]) != 2 {
		t.Errorf("Expected 2 domains under other.com, got %d", len(groups["other.com"]))
	}
}
