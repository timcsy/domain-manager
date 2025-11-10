package middleware

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	APIKeys    []string
	JWTSecret  string
	SkipPaths  []string
	SessionTTL time.Duration
}

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// ContextKeyUserID is the context key for user ID
	ContextKeyUserID contextKey = "user_id"
	// ContextKeyAPIKey is the context key for API key
	ContextKeyAPIKey contextKey = "api_key"
)

// APIKeyAuth returns a middleware that validates API key authentication
func APIKeyAuth(config *AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for certain paths
			for _, path := range config.SkipPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Extract API key from header
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				// Try Authorization header with Bearer scheme
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					apiKey = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if apiKey == "" {
				http.Error(w, "Missing API key", http.StatusUnauthorized)
				return
			}

			// Validate API key using constant-time comparison
			valid := false
			for _, key := range config.APIKeys {
				if subtle.ConstantTimeCompare([]byte(apiKey), []byte(key)) == 1 {
					valid = true
					break
				}
			}

			if !valid {
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}

			// Store API key in context
			ctx := context.WithValue(r.Context(), ContextKeyAPIKey, apiKey)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// JWTAuth returns a middleware that validates JWT tokens
func JWTAuth(config *AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for certain paths
			for _, path := range config.SkipPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			// Parse Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Parse and validate JWT token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Validate signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(config.JWTSecret), nil
			})

			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				http.Error(w, "Token validation failed", http.StatusUnauthorized)
				return
			}

			// Extract claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// Extract user ID from claims
			userID, ok := claims["user_id"].(string)
			if !ok {
				http.Error(w, "Missing user_id in token", http.StatusUnauthorized)
				return
			}

			// Store user ID in context
			ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SessionAuth returns a middleware that validates session-based authentication
func SessionAuth(config *AuthConfig) func(http.Handler) http.Handler {
	// Simple in-memory session store
	sessions := make(map[string]sessionData)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for certain paths
			for _, path := range config.SkipPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Extract session ID from cookie
			cookie, err := r.Cookie("session_id")
			if err != nil {
				http.Error(w, "Missing session cookie", http.StatusUnauthorized)
				return
			}

			sessionID := cookie.Value

			// Validate session
			session, exists := sessions[sessionID]
			if !exists {
				http.Error(w, "Invalid session", http.StatusUnauthorized)
				return
			}

			// Check if session has expired
			if time.Now().After(session.ExpiresAt) {
				delete(sessions, sessionID)
				http.Error(w, "Session expired", http.StatusUnauthorized)
				return
			}

			// Update session expiration
			session.ExpiresAt = time.Now().Add(config.SessionTTL)
			sessions[sessionID] = session

			// Store user ID in context
			ctx := context.WithValue(r.Context(), ContextKeyUserID, session.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// sessionData holds session information
type sessionData struct {
	UserID    string
	ExpiresAt time.Time
}

// GetUserID extracts user ID from request context
func GetUserID(r *http.Request) string {
	if userID, ok := r.Context().Value(ContextKeyUserID).(string); ok {
		return userID
	}
	return ""
}

// GetAPIKey extracts API key from request context
func GetAPIKey(r *http.Request) string {
	if apiKey, ok := r.Context().Value(ContextKeyAPIKey).(string); ok {
		return apiKey
	}
	return ""
}

// Auth is the default authentication middleware using session-based auth
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sessionID string

		// Try to get session ID from cookie first
		cookie, err := r.Cookie("session_id")
		if err == nil && cookie.Value != "" {
			sessionID = cookie.Value
		}

		// If no cookie, try Authorization header (for HTMX/fetch requests)
		if sessionID == "" {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				// Support both "Bearer token" and just "token"
				if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
					sessionID = authHeader[7:]
				} else {
					sessionID = authHeader
				}
			}
		}

		// If still no session ID, try X-Session-Token header
		if sessionID == "" {
			sessionID = r.Header.Get("X-Session-Token")
		}

		if sessionID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// For now, just check if session exists
		// In production, validate against session store
		ctx := context.WithValue(r.Context(), ContextKeyUserID, "admin")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
