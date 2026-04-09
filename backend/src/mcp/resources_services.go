package mcp

import (
	"encoding/json"

	"github.com/domain-manager/backend/src/k8s"
)

func (s *Server) registerServiceResources() {
	s.resourceList = append(s.resourceList,
		Resource{
			URI:         "service://list",
			Name:        "Service List",
			Description: "List of all Kubernetes services available for domain mapping",
			MimeType:    "application/json",
		},
	)

	s.resources["service://list"] = s.handleServiceListResource
}

func (s *Server) handleServiceListResource(uri string) (*ResourceResult, error) {
	mgr := k8s.NewServiceManager()
	services, err := mgr.ListAllServices()
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(services, "", "  ")
	if err != nil {
		return nil, err
	}

	return &ResourceResult{
		Contents: []ResourceContent{{
			URI:      uri,
			MimeType: "application/json",
			Text:     string(data),
		}},
	}, nil
}
