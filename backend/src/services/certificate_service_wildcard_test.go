package services

import (
	"testing"
)

// TestMatchesWildcard tests the wildcard matching logic
func TestMatchesWildcard(t *testing.T) {
	service := &CertificateService{}

	tests := []struct {
		name     string
		pattern  string
		domain   string
		expected bool
	}{
		{
			name:     "Wildcard matches single-level subdomain",
			pattern:  "*.example.com",
			domain:   "api.example.com",
			expected: true,
		},
		{
			name:     "Wildcard does not match root domain",
			pattern:  "*.example.com",
			domain:   "example.com",
			expected: false,
		},
		{
			name:     "Wildcard does not match multi-level subdomain",
			pattern:  "*.example.com",
			domain:   "sub.api.example.com",
			expected: false,
		},
		{
			name:     "Wildcard matches another single-level subdomain",
			pattern:  "*.example.com",
			domain:   "www.example.com",
			expected: true,
		},
		{
			name:     "Non-wildcard pattern returns false",
			pattern:  "example.com",
			domain:   "api.example.com",
			expected: false,
		},
		{
			name:     "Different root domain",
			pattern:  "*.example.com",
			domain:   "api.other.com",
			expected: false,
		},
		{
			name:     "Wildcard with different TLD",
			pattern:  "*.example.org",
			domain:   "api.example.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.matchesWildcard(tt.pattern, tt.domain)
			if result != tt.expected {
				t.Errorf("matchesWildcard(%q, %q) = %v; expected %v", tt.pattern, tt.domain, result, tt.expected)
			}
		})
	}
}

// TestEndsWith tests the endsWith helper function
func TestEndsWith(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		suffix   string
		expected bool
	}{
		{
			name:     "String ends with suffix",
			s:        "api.example.com",
			suffix:   "example.com",
			expected: true,
		},
		{
			name:     "String does not end with suffix",
			s:        "api.example.com",
			suffix:   "other.com",
			expected: false,
		},
		{
			name:     "Suffix longer than string",
			s:        "api.com",
			suffix:   "example.api.com",
			expected: false,
		},
		{
			name:     "Empty suffix",
			s:        "api.example.com",
			suffix:   "",
			expected: true,
		},
		{
			name:     "Both empty",
			s:        "",
			suffix:   "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := endsWith(tt.s, tt.suffix)
			if result != tt.expected {
				t.Errorf("endsWith(%q, %q) = %v; expected %v", tt.s, tt.suffix, result, tt.expected)
			}
		})
	}
}
