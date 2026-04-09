package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const (
	// ContextKeyRequestID is the context key for request ID
	ContextKeyRequestID contextKey = "request_id"

	// RequestIDHeader is the header name for request ID
	RequestIDHeader = "X-Request-ID"
)

// RequestTracing adds a unique request ID to each request for distributed tracing
func RequestTracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use existing request ID if provided, otherwise generate one
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Set response header
		w.Header().Set(RequestIDHeader, requestID)

		// Store in context
		ctx := context.WithValue(r.Context(), ContextKeyRequestID, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID extracts request ID from context
func GetRequestID(r *http.Request) string {
	if id, ok := r.Context().Value(ContextKeyRequestID).(string); ok {
		return id
	}
	return ""
}

func generateRequestID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
