package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// Logger is the logging middleware
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Generate request ID
		requestID := middleware.GetReqID(r.Context())

		// Wrap response writer to capture status code
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		// Log request
		log.Printf("[%s] %s %s %s - Started", requestID, r.RemoteAddr, r.Method, r.URL.Path)

		// Process request
		next.ServeHTTP(ww, r)

		// Log response
		duration := time.Since(start)
		log.Printf("[%s] %s %s %s - Completed %d in %v",
			requestID, r.RemoteAddr, r.Method, r.URL.Path, ww.Status(), duration)
	})
}
