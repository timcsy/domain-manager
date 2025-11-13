package subdomain

import (
	"fmt"
	"regexp"
	"strings"
)

// SubdomainValidator provides subdomain validation functionality
type SubdomainValidator struct{}

// NewValidator creates a new subdomain validator
func NewValidator() *SubdomainValidator {
	return &SubdomainValidator{}
}

// ValidationResult contains the results of subdomain validation
type ValidationResult struct {
	Valid        bool
	IsSubdomain  bool
	RootDomain   string
	Subdomain    string
	ErrorMessage string
}

// ValidateSubdomain validates a subdomain format and structure
func (v *SubdomainValidator) ValidateSubdomain(domain string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid: true,
	}

	// Trim whitespace
	domain = strings.TrimSpace(domain)

	// Check if domain is empty
	if domain == "" {
		result.Valid = false
		result.ErrorMessage = "domain cannot be empty"
		return result, nil
	}

	// Check domain format (RFC 1035/1123 compliant)
	domainRegex := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
	if !domainRegex.MatchString(domain) {
		result.Valid = false
		result.ErrorMessage = "invalid domain format"
		return result, nil
	}

	// Check label length (each label must be <= 63 chars)
	labels := strings.Split(domain, ".")
	for _, label := range labels {
		if len(label) > 63 {
			result.Valid = false
			result.ErrorMessage = fmt.Sprintf("label '%s' exceeds 63 characters", label)
			return result, nil
		}
		if len(label) == 0 {
			result.Valid = false
			result.ErrorMessage = "empty label in domain"
			return result, nil
		}
	}

	// Check total length (must be <= 253 chars)
	if len(domain) > 253 {
		result.Valid = false
		result.ErrorMessage = "domain exceeds 253 characters"
		return result, nil
	}

	// Determine if it's a subdomain
	if len(labels) > 2 {
		result.IsSubdomain = true
		// Extract root domain (last two labels)
		result.RootDomain = strings.Join(labels[len(labels)-2:], ".")
		// Extract subdomain part
		result.Subdomain = strings.Join(labels[:len(labels)-2], ".")
	} else {
		result.IsSubdomain = false
		result.RootDomain = domain
	}

	return result, nil
}

// GetRootDomain extracts the root domain from a domain name
func (v *SubdomainValidator) GetRootDomain(domain string) string {
	domain = strings.TrimSpace(domain)
	labels := strings.Split(domain, ".")

	if len(labels) >= 2 {
		return strings.Join(labels[len(labels)-2:], ".")
	}

	return domain
}

// GetSubdomainPart extracts the subdomain part from a domain name
func (v *SubdomainValidator) GetSubdomainPart(domain string) string {
	domain = strings.TrimSpace(domain)
	labels := strings.Split(domain, ".")

	if len(labels) > 2 {
		return strings.Join(labels[:len(labels)-2], ".")
	}

	return ""
}

// IsSubdomain checks if a domain is a subdomain
func (v *SubdomainValidator) IsSubdomain(domain string) bool {
	domain = strings.TrimSpace(domain)
	labels := strings.Split(domain, ".")
	return len(labels) > 2
}

// SubdomainConflict represents a subdomain conflict
type SubdomainConflict struct {
	ConflictType   string // "exact", "parent", "child", "wildcard"
	ConflictDomain string
	Message        string
}

// CheckSubdomainConflict checks for conflicts with existing domains
func (v *SubdomainValidator) CheckSubdomainConflict(newDomain string, existingDomains []string) []SubdomainConflict {
	conflicts := []SubdomainConflict{}

	newDomain = strings.TrimSpace(strings.ToLower(newDomain))

	for _, existing := range existingDomains {
		existing = strings.TrimSpace(strings.ToLower(existing))

		// Skip if same domain
		if newDomain == existing {
			conflicts = append(conflicts, SubdomainConflict{
				ConflictType:   "exact",
				ConflictDomain: existing,
				Message:        fmt.Sprintf("domain '%s' already exists", existing),
			})
			continue
		}

		// Check if new domain is a subdomain of existing
		if v.isChildOf(newDomain, existing) {
			conflicts = append(conflicts, SubdomainConflict{
				ConflictType:   "parent",
				ConflictDomain: existing,
				Message:        fmt.Sprintf("'%s' is a subdomain of existing domain '%s'", newDomain, existing),
			})
		}

		// Check if existing domain is a subdomain of new
		if v.isChildOf(existing, newDomain) {
			conflicts = append(conflicts, SubdomainConflict{
				ConflictType:   "child",
				ConflictDomain: existing,
				Message:        fmt.Sprintf("existing domain '%s' is a subdomain of '%s'", existing, newDomain),
			})
		}

		// Check wildcard conflicts
		if strings.HasPrefix(existing, "*.") {
			wildcardRoot := existing[2:] // Remove "*."
			newRoot := v.GetRootDomain(newDomain)
			if wildcardRoot == newRoot || strings.HasSuffix(newDomain, "."+wildcardRoot) {
				conflicts = append(conflicts, SubdomainConflict{
					ConflictType:   "wildcard",
					ConflictDomain: existing,
					Message:        fmt.Sprintf("'%s' conflicts with wildcard domain '%s'", newDomain, existing),
				})
			}
		}

		if strings.HasPrefix(newDomain, "*.") {
			wildcardRoot := newDomain[2:] // Remove "*."
			existingRoot := v.GetRootDomain(existing)
			if wildcardRoot == existingRoot || strings.HasSuffix(existing, "."+wildcardRoot) {
				conflicts = append(conflicts, SubdomainConflict{
					ConflictType:   "wildcard",
					ConflictDomain: existing,
					Message:        fmt.Sprintf("wildcard domain '%s' conflicts with existing domain '%s'", newDomain, existing),
				})
			}
		}
	}

	return conflicts
}

// isChildOf checks if child is a subdomain of parent
func (v *SubdomainValidator) isChildOf(child, parent string) bool {
	// child must end with ".parent" and be longer
	if !strings.HasSuffix(child, "."+parent) {
		return false
	}

	// Ensure it's a direct or indirect subdomain, not just a suffix match
	return len(child) > len(parent)+1
}

// GroupByRootDomain groups domains by their root domain
func (v *SubdomainValidator) GroupByRootDomain(domains []string) map[string][]string {
	groups := make(map[string][]string)

	for _, domain := range domains {
		root := v.GetRootDomain(domain)
		groups[root] = append(groups[root], domain)
	}

	return groups
}
