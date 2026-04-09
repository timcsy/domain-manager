package mcp

import (
	"encoding/json"
	"strings"

	"github.com/domain-manager/backend/src/models"
)

func (s *Server) registerDomainResources() {
	s.resourceList = append(s.resourceList,
		Resource{
			URI:         "domain://list",
			Name:        "Domain List",
			Description: "List of all configured domains",
			MimeType:    "application/json",
		},
		Resource{
			URI:         "domain://{name}",
			Name:        "Domain Details",
			Description: "Details of a specific domain by name",
			MimeType:    "application/json",
		},
	)

	s.resources["domain://list"] = s.handleDomainListResource
	s.resources["domain://{name}"] = s.handleDomainDetailResource
}

func (s *Server) handleDomainListResource(uri string) (*ResourceResult, error) {
	domains, _, err := s.domainService.ListDomains(models.DomainFilter{})
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(domains, "", "  ")
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

func (s *Server) handleDomainDetailResource(uri string) (*ResourceResult, error) {
	// Extract domain name from URI: domain://example.com
	name := strings.TrimPrefix(uri, "domain://")

	domain, err := s.domainRepo.GetByName(name)
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(domain, "", "  ")
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
