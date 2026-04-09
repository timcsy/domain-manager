package mcp

import (
	"encoding/json"
)

func (s *Server) registerCertificateResources() {
	s.resourceList = append(s.resourceList,
		Resource{
			URI:         "certificate://list",
			Name:        "Certificate List",
			Description: "List of all SSL certificates",
			MimeType:    "application/json",
		},
	)

	s.resources["certificate://list"] = s.handleCertificateListResource
}

func (s *Server) handleCertificateListResource(uri string) (*ResourceResult, error) {
	certs, _, err := s.certificateService.ListCertificates(100, 0)
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(certs, "", "  ")
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
