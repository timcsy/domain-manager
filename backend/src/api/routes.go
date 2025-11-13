package api

import (
	"net/http"
	"os"

	"github.com/domain-manager/backend/src/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// NewRouter creates and configures the API router
func NewRouter() *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(chimiddleware.Compress(5))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key", "X-Session-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check (public)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		Success(w, map[string]string{"status": "healthy"}, "Service is running")
	})

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Group(func(r chi.Router) {
			r.Post("/auth/login", HandleLogin)
			r.Post("/auth/logout", HandleLogout)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth)

			// Domains - literal routes must come before parameter routes
			r.Get("/domains", HandleListDomains)
			r.Post("/domains", HandleCreateDomain)
			r.Get("/domains/tree", HandleGetDomainTree)
			r.Post("/domains/batch", HandleBatchCreateDomains)
		r.Delete("/domains/batch", HandleBatchDeleteDomains)
		r.Patch("/domains/batch", HandleBatchUpdateDomains)
			r.Get("/domains/{id}", HandleGetDomain)
			r.Put("/domains/{id}", HandleUpdateDomain)
			r.Delete("/domains/{id}", HandleDeleteDomain)
			r.Get("/domains/{id}/status", HandleGetDomainStatus)
			r.Get("/domains/{id}/diagnostics", HandleGetDomainDiagnostics)

			// Certificates
			r.Route("/certificates", func(r chi.Router) {
				r.Get("/", HandleListCertificates)
				r.Post("/", HandleUploadCertificate)
				r.Get("/expiring", HandleGetExpiringCertificates)
				r.Get("/{id}", HandleGetCertificate)
				r.Delete("/{id}", HandleDeleteCertificate)
				r.Post("/{id}/renew", HandleRenewCertificate)
			})

			// Services (Kubernetes)
			r.Route("/services", func(r chi.Router) {
				r.Get("/", HandleListServices)
				r.Get("/{namespace}/{name}", HandleGetService)
			})

			// Diagnostics
			r.Route("/diagnostics", func(r chi.Router) {
				r.Get("/health", HandleHealthCheck)
				r.Get("/logs", HandleGetLogs)
			})

			// Settings
			r.Route("/settings", func(r chi.Router) {
				r.Get("/", HandleGetSettings)
				r.Patch("/", HandleUpdateSettings)
			})

			// API Keys
			r.Route("/api-keys", func(r chi.Router) {
				r.Get("/", HandleListAPIKeys)
				r.Post("/", HandleCreateAPIKey)
				r.Delete("/{id}", HandleDeleteAPIKey)
			})
		})
	})

	// MCP endpoint
	r.Post("/mcp", HandleMCP)

	// Serve frontend static files
	frontendDir := os.Getenv("FRONTEND_PATH")
	if frontendDir == "" {
		frontendDir = "../frontend" // default for local development
	}
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, frontendDir+"/src/pages/login.html")
	})
	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, frontendDir+"/src/pages/login.html")
	})
	r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, frontendDir+"/src/pages/dashboard.html")
	})
	r.Get("/domains", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, frontendDir+"/src/pages/domains.html")
	})
	r.Get("/domain-detail", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, frontendDir+"/src/pages/domain-detail.html")
	})
	r.Get("/certificates", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, frontendDir+"/src/pages/certificates.html")
	})

	// Serve static assets with correct MIME types
	r.Get("/dist/styles/*", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, frontendDir+r.URL.Path)
	})
	r.Get("/dist/js/*", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, frontendDir+r.URL.Path)
	})

	// Serve component partials
	r.Get("/components/*", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, frontendDir+"/src"+r.URL.Path)
	})

	return r
}

// Additional route for domain status
func init() {
	// Initialize services when package is loaded
	// This will be called after database is initialized
}
