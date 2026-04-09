package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/domain-manager/backend/src/db"
	"github.com/domain-manager/backend/src/k8s"
	"github.com/domain-manager/backend/src/models"
	"github.com/domain-manager/backend/src/repositories"
)

func (s *Server) registerDiagnosticTools() {
	s.toolList = append(s.toolList,
		Tool{
			Name:        "check_dns",
			Description: "Check DNS resolution for a domain",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"domain": map[string]string{"type": "string", "description": "Domain name to check"},
				},
				Required: []string{"domain"},
			},
		},
		Tool{
			Name:        "get_diagnostics",
			Description: "Get diagnostic logs with optional filters",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"category": map[string]string{"type": "string", "description": "Filter by check type: health, ssl, dns, connectivity"},
					"log_type": map[string]string{"type": "string", "description": "Filter by status: success, warning, error"},
					"limit":    map[string]string{"type": "number", "description": "Number of logs to return (default: 50)"},
				},
			},
		},
		Tool{
			Name:        "get_system_health",
			Description: "Get overall system health status including database, Kubernetes, and certificate status",
			InputSchema: InputSchema{
				Type: "object",
			},
		},
	)

	s.tools["check_dns"] = s.handleCheckDNS
	s.tools["get_diagnostics"] = s.handleGetDiagnostics
	s.tools["get_system_health"] = s.handleGetSystemHealth
}

func (s *Server) handleCheckDNS(params json.RawMessage) (*ToolResult, error) {
	var args struct {
		Domain string `json:"domain"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	result, err := s.domainService.ValidateSubdomain(args.Domain)
	if err != nil {
		return nil, err
	}
	return jsonResult(result)
}

func (s *Server) handleGetDiagnostics(params json.RawMessage) (*ToolResult, error) {
	var args struct {
		CheckType string  `json:"category"`
		Status    string  `json:"log_type"`
		Limit     float64 `json:"limit"`
	}
	if len(params) > 0 {
		_ = json.Unmarshal(params, &args)
	}
	if args.Limit == 0 {
		args.Limit = 50
	}

	diagRepo := repositories.NewDiagnosticLogRepository(db.DB)
	filter := models.DiagnosticLogFilter{
		CheckType: args.CheckType,
		Status:    args.Status,
		Limit:     int(args.Limit),
	}

	logs, err := diagRepo.List(filter)
	if err != nil {
		return nil, err
	}
	return jsonResult(logs)
}

func (s *Server) handleGetSystemHealth(params json.RawMessage) (*ToolResult, error) {
	health := map[string]interface{}{
		"database":   "healthy",
		"kubernetes": "unknown",
	}

	// Check database
	if err := db.Health(); err != nil {
		health["database"] = fmt.Sprintf("unhealthy: %v", err)
	}

	// Check Kubernetes
	mgr := k8s.NewServiceManager()
	if _, err := mgr.ListNamespaces(); err != nil {
		health["kubernetes"] = fmt.Sprintf("unhealthy: %v", err)
	} else {
		health["kubernetes"] = "healthy"
	}

	// Get domain stats
	domains, total, _ := s.domainService.ListDomains(models.DomainFilter{})
	activeCount := 0
	for _, d := range domains {
		if d.Status == "active" && d.Enabled {
			activeCount++
		}
	}
	health["domains"] = map[string]int{
		"total":  total,
		"active": activeCount,
	}

	// Get expiring certificates
	expiringCerts, _ := s.certificateService.GetExpiringCertificates(30)
	health["certificates"] = map[string]interface{}{
		"expiring_within_30_days": len(expiringCerts),
	}

	return jsonResult(health)
}
