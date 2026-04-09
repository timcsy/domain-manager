package mcp

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/domain-manager/backend/src/repositories"
	"github.com/domain-manager/backend/src/services"
)

// JSON-RPC 2.0 types
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// MCP protocol types
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	ProtocolVersion string            `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo        `json:"serverInfo"`
}

type ServerCapabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
}

type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
}

type ToolResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
}

type ResourceResult struct {
	Contents []ResourceContent `json:"contents"`
}

// Server is the MCP server
type Server struct {
	domainService      *services.DomainService
	certificateService *services.CertificateService
	domainRepo         *repositories.DomainRepository
	certRepo           *repositories.CertificateRepository
	tools              map[string]ToolHandler
	resources          map[string]ResourceHandler
	toolList           []Tool
	resourceList       []Resource
}

type ToolHandler func(params json.RawMessage) (*ToolResult, error)
type ResourceHandler func(uri string) (*ResourceResult, error)

// NewServer creates a new MCP server
func NewServer(
	domainService *services.DomainService,
	certificateService *services.CertificateService,
	domainRepo *repositories.DomainRepository,
	certRepo *repositories.CertificateRepository,
) *Server {
	s := &Server{
		domainService:      domainService,
		certificateService: certificateService,
		domainRepo:         domainRepo,
		certRepo:           certRepo,
		tools:              make(map[string]ToolHandler),
		resources:          make(map[string]ResourceHandler),
	}
	s.registerTools()
	s.registerResources()
	return s
}

func (s *Server) registerTools() {
	s.registerDomainTools()
	s.registerServiceTools()
	s.registerCertificateTools()
	s.registerDiagnosticTools()
}

func (s *Server) registerResources() {
	s.registerDomainResources()
	s.registerServiceResources()
	s.registerCertificateResources()
	s.registerDiagnosticResources()
}

// HandleMessage processes a JSON-RPC 2.0 message and returns a response
func (s *Server) HandleMessage(data []byte) []byte {
	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return s.errorResponse(nil, NewParseError(err.Error()))
	}

	if req.JSONRPC != "2.0" {
		return s.errorResponse(req.ID, NewInvalidRequest("jsonrpc must be 2.0"))
	}

	log.Printf("MCP request: method=%s", req.Method)

	var result interface{}
	var rpcErr *RPCError

	switch req.Method {
	case "initialize":
		result = s.handleInitialize()
	case "initialized":
		// Notification, no response needed
		return nil
	case "tools/list":
		result = s.handleToolsList()
	case "tools/call":
		result, rpcErr = s.handleToolsCall(req.Params)
	case "resources/list":
		result = s.handleResourcesList()
	case "resources/read":
		result, rpcErr = s.handleResourcesRead(req.Params)
	case "ping":
		result = map[string]interface{}{}
	default:
		rpcErr = NewMethodNotFound(req.Method)
	}

	if rpcErr != nil {
		return s.errorResponse(req.ID, rpcErr)
	}

	return s.successResponse(req.ID, result)
}

func (s *Server) handleInitialize() *InitializeResult {
	return &InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapabilities{
			Tools:     &ToolsCapability{},
			Resources: &ResourcesCapability{},
		},
		ServerInfo: ServerInfo{
			Name:    "domain-manager",
			Version: "1.0.0",
		},
	}
}

func (s *Server) handleToolsList() map[string]interface{} {
	return map[string]interface{}{
		"tools": s.toolList,
	}
}

func (s *Server) handleToolsCall(params json.RawMessage) (*ToolResult, *RPCError) {
	var callParams struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(params, &callParams); err != nil {
		return nil, NewInvalidParams(err.Error())
	}

	handler, ok := s.tools[callParams.Name]
	if !ok {
		return nil, NewMethodNotFound(callParams.Name)
	}

	result, err := handler(callParams.Arguments)
	if err != nil {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
			IsError: true,
		}, nil
	}

	return result, nil
}

func (s *Server) handleResourcesList() map[string]interface{} {
	return map[string]interface{}{
		"resources": s.resourceList,
	}
}

func (s *Server) handleResourcesRead(params json.RawMessage) (*ResourceResult, *RPCError) {
	var readParams struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(params, &readParams); err != nil {
		return nil, NewInvalidParams(err.Error())
	}

	handler, ok := s.resources[readParams.URI]
	if !ok {
		// Try prefix matching for parameterized resources
		for pattern, h := range s.resources {
			if matchResourceURI(pattern, readParams.URI) {
				handler = h
				ok = true
				break
			}
		}
	}

	if !ok {
		return nil, NewInvalidParams("resource not found: " + readParams.URI)
	}

	result, err := handler(readParams.URI)
	if err != nil {
		return nil, NewInternalError(err.Error())
	}

	return result, nil
}

func (s *Server) successResponse(id interface{}, result interface{}) []byte {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	data, _ := json.Marshal(resp)
	return data
}

func (s *Server) errorResponse(id interface{}, rpcErr *RPCError) []byte {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Error:   rpcErr,
	}
	data, _ := json.Marshal(resp)
	return data
}

// textResult is a helper to create a simple text tool result
func textResult(text string) *ToolResult {
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: text}},
	}
}

// jsonResult is a helper to create a JSON tool result
func jsonResult(v interface{}) (*ToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return textResult(string(data)), nil
}

// matchResourceURI checks if a URI matches a pattern like "domain://{name}"
func matchResourceURI(pattern, uri string) bool {
	// Simple prefix match for patterns with parameters
	if len(pattern) > 0 && pattern[len(pattern)-1] == '}' {
		idx := len(pattern) - 1
		for idx > 0 && pattern[idx] != '{' {
			idx--
		}
		if idx > 0 {
			prefix := pattern[:idx]
			return len(uri) > len(prefix) && uri[:len(prefix)] == prefix
		}
	}
	return false
}
