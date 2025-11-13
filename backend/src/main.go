package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/domain-manager/backend/src/api"
	"github.com/domain-manager/backend/src/db"
	"github.com/domain-manager/backend/src/k8s"
	"github.com/domain-manager/backend/src/services/scheduler"
)

func main() {
	log.Println("Starting Kubernetes Domain Manager...")

	// Initialize database
	dbConfig := db.DefaultConfig()
	if err := db.Initialize(dbConfig); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize Kubernetes client
	k8sConfig := k8s.DefaultConfig()
	if err := k8s.Initialize(k8sConfig); err != nil {
		log.Fatalf("Failed to initialize Kubernetes client: %v", err)
	}

	// Initialize services
	api.InitializeServices()

	// Initialize and start scheduler
	if err := api.InitializeScheduler(); err != nil {
		log.Fatalf("Failed to initialize scheduler: %v", err)
	}

	// Create router
	router := api.NewRouter()

	// Configure HTTP server
	port := getEnv("PORT", "8080")
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop scheduler
	log.Println("Stopping scheduler...")
	if err := scheduler.StopGlobal(); err != nil {
		log.Printf("Warning: Failed to stop scheduler: %v", err)
	}

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited successfully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
