package mcp

import (
	"encoding/json"

	"github.com/domain-manager/backend/src/db"
	"github.com/domain-manager/backend/src/models"
	"github.com/domain-manager/backend/src/repositories"
)

func (s *Server) registerDiagnosticResources() {
	s.resourceList = append(s.resourceList,
		Resource{
			URI:         "diagnostics://logs",
			Name:        "Diagnostic Logs",
			Description: "Recent diagnostic logs from the system",
			MimeType:    "application/json",
		},
	)

	s.resources["diagnostics://logs"] = s.handleDiagnosticLogsResource
}

func (s *Server) handleDiagnosticLogsResource(uri string) (*ResourceResult, error) {
	diagRepo := repositories.NewDiagnosticLogRepository(db.DB)
	logs, err := diagRepo.List(models.DiagnosticLogFilter{
		Limit: 100,
	})
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(logs, "", "  ")
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
