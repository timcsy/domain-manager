package models

import (
	"time"
)

// Domain represents a domain configuration
type Domain struct {
	ID              int64     `json:"id" db:"id"`
	DomainName      string    `json:"domain_name" db:"domain_name"`
	TargetService   string    `json:"target_service" db:"target_service"`
	TargetNamespace string    `json:"target_namespace" db:"target_namespace"`
	TargetPort      int       `json:"target_port" db:"target_port"`
	SSLMode         string    `json:"ssl_mode" db:"ssl_mode"` // "auto" or "manual"
	CertificateID   *int64    `json:"certificate_id,omitempty" db:"certificate_id"`
	Status          string    `json:"status" db:"status"` // "pending", "active", "error", "deleted"
	Enabled         bool      `json:"enabled" db:"enabled"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// DomainCreateRequest represents the request to create a domain
type DomainCreateRequest struct {
	DomainName      string `json:"domain_name" validate:"required,fqdn"`
	TargetService   string `json:"target_service" validate:"required"`
	TargetNamespace string `json:"target_namespace"`
	TargetPort      int    `json:"target_port" validate:"required,min=1,max=65535"`
	SSLMode         string `json:"ssl_mode"` // defaults to "auto"
}

// DomainUpdateRequest represents the request to update a domain
type DomainUpdateRequest struct {
	TargetService   *string `json:"target_service,omitempty"`
	TargetNamespace *string `json:"target_namespace,omitempty"`
	TargetPort      *int    `json:"target_port,omitempty"`
	SSLMode         *string `json:"ssl_mode,omitempty"`
	Status          *string `json:"status,omitempty"`
	Enabled         *bool   `json:"enabled,omitempty"`
}

// DomainFilter represents filters for listing domains
type DomainFilter struct {
	Status      string
	Enabled     *bool
	ServiceName string
	Namespace   string
	Limit       int
	Offset      int
}

// Validate validates the domain create request
func (r *DomainCreateRequest) Validate() error {
	if r.DomainName == "" {
		return ErrInvalidDomainName
	}
	if r.TargetService == "" {
		return ErrInvalidServiceName
	}
	if r.TargetPort < 1 || r.TargetPort > 65535 {
		return ErrInvalidPort
	}
	if r.TargetNamespace == "" {
		r.TargetNamespace = "default"
	}
	if r.SSLMode == "" {
		r.SSLMode = "auto"
	}
	return nil
}

// DomainTreeNode represents a node in the domain tree structure
type DomainTreeNode struct {
	Domain      *Domain            `json:"domain"`
	RootDomain  string             `json:"root_domain"`
	Subdomains  []*DomainTreeNode  `json:"subdomains,omitempty"`
	Count       int                `json:"count"` // Total count including nested subdomains
}
