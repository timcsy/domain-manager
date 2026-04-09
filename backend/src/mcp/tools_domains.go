package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/domain-manager/backend/src/models"
)

func (s *Server) registerDomainTools() {
	s.toolList = append(s.toolList,
		Tool{
			Name:        "list_domains",
			Description: "List all domains with optional filters",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"status":  map[string]string{"type": "string", "description": "Filter by status: pending, active, error, deleted"},
					"service": map[string]string{"type": "string", "description": "Filter by target service name"},
				},
			},
		},
		Tool{
			Name:        "get_domain",
			Description: "Get details of a specific domain by ID",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"id": map[string]string{"type": "number", "description": "Domain ID"},
				},
				Required: []string{"id"},
			},
		},
		Tool{
			Name:        "create_domain",
			Description: "Create a new domain mapping to a Kubernetes service",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"domain_name":      map[string]string{"type": "string", "description": "FQDN of the domain"},
					"target_service":   map[string]string{"type": "string", "description": "Kubernetes service name"},
					"target_namespace": map[string]string{"type": "string", "description": "Kubernetes namespace (default: default)"},
					"target_port":      map[string]string{"type": "number", "description": "Target port (1-65535)"},
					"ssl_mode":         map[string]string{"type": "string", "description": "SSL mode: auto or manual (default: auto)"},
				},
				Required: []string{"domain_name", "target_service", "target_port"},
			},
		},
		Tool{
			Name:        "update_domain",
			Description: "Update an existing domain",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"id":               map[string]string{"type": "number", "description": "Domain ID"},
					"target_service":   map[string]string{"type": "string", "description": "New target service"},
					"target_namespace": map[string]string{"type": "string", "description": "New target namespace"},
					"target_port":      map[string]string{"type": "number", "description": "New target port"},
					"ssl_mode":         map[string]string{"type": "string", "description": "New SSL mode"},
					"enabled":          map[string]string{"type": "boolean", "description": "Enable/disable domain"},
				},
				Required: []string{"id"},
			},
		},
		Tool{
			Name:        "delete_domain",
			Description: "Delete a domain by ID",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"id":   map[string]string{"type": "number", "description": "Domain ID"},
					"hard": map[string]string{"type": "boolean", "description": "Hard delete (default: false, soft delete)"},
				},
				Required: []string{"id"},
			},
		},
	)

	s.tools["list_domains"] = s.handleListDomains
	s.tools["get_domain"] = s.handleGetDomain
	s.tools["create_domain"] = s.handleCreateDomain
	s.tools["update_domain"] = s.handleUpdateDomain
	s.tools["delete_domain"] = s.handleDeleteDomain
}

func (s *Server) handleListDomains(params json.RawMessage) (*ToolResult, error) {
	var args struct {
		Status  string `json:"status"`
		Service string `json:"service"`
	}
	if len(params) > 0 {
		json.Unmarshal(params, &args)
	}

	filter := models.DomainFilter{
		Status:      args.Status,
		ServiceName: args.Service,
	}
	domains, total, err := s.domainService.ListDomains(filter)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"domains": domains,
		"total":   total,
	}
	return jsonResult(result)
}

func (s *Server) handleGetDomain(params json.RawMessage) (*ToolResult, error) {
	var args struct {
		ID float64 `json:"id"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	domain, err := s.domainService.GetDomainByID(int64(args.ID))
	if err != nil {
		return nil, err
	}
	return jsonResult(domain)
}

func (s *Server) handleCreateDomain(params json.RawMessage) (*ToolResult, error) {
	var req models.DomainCreateRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	domain, err := s.domainService.CreateDomain(&req)
	if err != nil {
		return nil, err
	}
	return jsonResult(domain)
}

func (s *Server) handleUpdateDomain(params json.RawMessage) (*ToolResult, error) {
	var args struct {
		ID              float64 `json:"id"`
		TargetService   *string `json:"target_service,omitempty"`
		TargetNamespace *string `json:"target_namespace,omitempty"`
		TargetPort      *int    `json:"target_port,omitempty"`
		SSLMode         *string `json:"ssl_mode,omitempty"`
		Enabled         *bool   `json:"enabled,omitempty"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	req := &models.DomainUpdateRequest{
		TargetService:   args.TargetService,
		TargetNamespace: args.TargetNamespace,
		TargetPort:      args.TargetPort,
		SSLMode:         args.SSLMode,
		Enabled:         args.Enabled,
	}

	domain, err := s.domainService.UpdateDomain(int64(args.ID), req)
	if err != nil {
		return nil, err
	}
	return jsonResult(domain)
}

func (s *Server) handleDeleteDomain(params json.RawMessage) (*ToolResult, error) {
	var args struct {
		ID   float64 `json:"id"`
		Hard bool    `json:"hard"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if err := s.domainService.DeleteDomain(int64(args.ID), args.Hard); err != nil {
		return nil, err
	}
	return textResult(fmt.Sprintf("Domain %d deleted successfully", int64(args.ID))), nil
}
