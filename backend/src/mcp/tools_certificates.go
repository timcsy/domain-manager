package mcp

import (
	"encoding/json"
	"fmt"
)

func (s *Server) registerCertificateTools() {
	s.toolList = append(s.toolList,
		Tool{
			Name:        "get_certificate_status",
			Description: "Get the SSL certificate status for a specific certificate",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"id": map[string]string{"type": "number", "description": "Certificate ID"},
				},
				Required: []string{"id"},
			},
		},
		Tool{
			Name:        "list_expiring_certificates",
			Description: "List certificates expiring within a given number of days",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"days": map[string]string{"type": "number", "description": "Number of days until expiration (default: 30)"},
				},
			},
		},
		Tool{
			Name:        "renew_certificate",
			Description: "Trigger renewal of a specific certificate",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"id": map[string]string{"type": "number", "description": "Certificate ID to renew"},
				},
				Required: []string{"id"},
			},
		},
	)

	s.tools["get_certificate_status"] = s.handleGetCertificateStatus
	s.tools["list_expiring_certificates"] = s.handleListExpiringCertificates
	s.tools["renew_certificate"] = s.handleRenewCertificate
}

func (s *Server) handleGetCertificateStatus(params json.RawMessage) (*ToolResult, error) {
	var args struct {
		ID float64 `json:"id"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	cert, err := s.certificateService.GetCertificateByID(int(args.ID))
	if err != nil {
		return nil, err
	}
	return jsonResult(cert)
}

func (s *Server) handleListExpiringCertificates(params json.RawMessage) (*ToolResult, error) {
	var args struct {
		Days float64 `json:"days"`
	}
	if len(params) > 0 {
		json.Unmarshal(params, &args)
	}
	if args.Days == 0 {
		args.Days = 30
	}

	certs, err := s.certificateService.GetExpiringCertificates(int(args.Days))
	if err != nil {
		return nil, err
	}
	return jsonResult(certs)
}

func (s *Server) handleRenewCertificate(params json.RawMessage) (*ToolResult, error) {
	var args struct {
		ID float64 `json:"id"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	cert, err := s.certificateService.RenewCertificate(int(args.ID))
	if err != nil {
		return nil, err
	}
	return jsonResult(cert)
}
