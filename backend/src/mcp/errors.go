package mcp

import "fmt"

// JSON-RPC 2.0 error codes
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
)

// RPCError represents a JSON-RPC 2.0 error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("RPC error %d: %s", e.Code, e.Message)
}

func NewParseError(msg string) *RPCError {
	return &RPCError{Code: CodeParseError, Message: "Parse error: " + msg}
}

func NewInvalidRequest(msg string) *RPCError {
	return &RPCError{Code: CodeInvalidRequest, Message: "Invalid request: " + msg}
}

func NewMethodNotFound(method string) *RPCError {
	return &RPCError{Code: CodeMethodNotFound, Message: "Method not found: " + method}
}

func NewInvalidParams(msg string) *RPCError {
	return &RPCError{Code: CodeInvalidParams, Message: "Invalid params: " + msg}
}

func NewInternalError(msg string) *RPCError {
	return &RPCError{Code: CodeInternalError, Message: "Internal error: " + msg}
}
