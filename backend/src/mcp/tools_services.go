package mcp

import (
	"encoding/json"

	"github.com/domain-manager/backend/src/k8s"
)

func (s *Server) registerServiceTools() {
	s.toolList = append(s.toolList,
		Tool{
			Name:        "list_services",
			Description: "List Kubernetes services available for domain mapping",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"namespace": map[string]string{"type": "string", "description": "Kubernetes namespace (empty for all namespaces)"},
				},
			},
		},
	)

	s.tools["list_services"] = s.handleListServices
}

func (s *Server) handleListServices(params json.RawMessage) (*ToolResult, error) {
	var args struct {
		Namespace string `json:"namespace"`
	}
	if len(params) > 0 {
		_ = json.Unmarshal(params, &args)
	}

	mgr := k8s.NewServiceManager()

	if args.Namespace == "" {
		services, err := mgr.ListAllServices()
		if err != nil {
			return nil, err
		}
		return jsonResult(services)
	}

	services, err := mgr.ListServices(args.Namespace)
	if err != nil {
		return nil, err
	}
	return jsonResult(services)
}
