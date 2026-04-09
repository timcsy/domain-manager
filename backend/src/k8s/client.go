package k8s

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Client is the global Kubernetes client
var Client *kubernetes.Clientset

// DynamicClient is the global Kubernetes dynamic client (for CRDs like cert-manager)
var DynamicClient dynamic.Interface

// Config holds Kubernetes client configuration
type Config struct {
	InCluster  bool
	Kubeconfig string
}

// DefaultConfig returns default Kubernetes configuration
func DefaultConfig() *Config {
	return &Config{
		InCluster:  getEnvBool("K8S_IN_CLUSTER", true),
		Kubeconfig: getEnv("KUBECONFIG", filepath.Join(homeDir(), ".kube", "config")),
	}
}

// Initialize initializes the Kubernetes client
func Initialize(cfg *Config) error {
	// Check if mock mode is enabled
	if getEnvBool("K8S_MOCK", false) {
		log.Println("🔧 K8S_MOCK=true detected - Using mock mode")
		return InitializeMock()
	}

	var config *rest.Config
	var err error

	if cfg.InCluster {
		// In-cluster configuration
		log.Println("Using in-cluster Kubernetes configuration")
		config, err = rest.InClusterConfig()
		if err != nil {
			// If in-cluster fails, offer to use mock mode
			log.Printf("⚠️  Failed to create in-cluster config: %v", err)
			log.Println("💡 Tip: Set K8S_MOCK=true to run without K8s cluster")
			return fmt.Errorf("failed to create in-cluster config (use K8S_MOCK=true for local dev): %w", err)
		}
	} else {
		// Kubeconfig file configuration
		log.Printf("Using kubeconfig from: %s", cfg.Kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", cfg.Kubeconfig)
		if err != nil {
			log.Printf("⚠️  Failed to load kubeconfig: %v", err)
			log.Println("💡 Tip: Set K8S_MOCK=true to run without K8s cluster")
			return fmt.Errorf("failed to build config from kubeconfig (use K8S_MOCK=true for local dev): %w", err)
		}
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	Client = clientset

	// Create dynamic client for CRDs (cert-manager ClusterIssuer etc.)
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}
	DynamicClient = dynClient

	log.Println("Kubernetes client initialized successfully")

	// Test connection
	if err := Health(); err != nil {
		log.Printf("⚠️  Failed to connect to K8s cluster: %v", err)
		log.Println("💡 Tip: Set K8S_MOCK=true to run without K8s cluster")
		return fmt.Errorf("failed to connect to Kubernetes cluster (use K8S_MOCK=true for local dev): %w", err)
	}

	return nil
}

// Health checks Kubernetes cluster connectivity
func Health() error {
	// Check if in mock mode
	if IsMockMode() {
		return HealthMock()
	}

	if Client == nil {
		return fmt.Errorf("Kubernetes client is not initialized")
	}

	// Try to list namespaces to verify connectivity
	ctx := context.Background()
	_, err := Client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to communicate with Kubernetes API: %w", err)
	}

	return nil
}

// GetClient returns the global Kubernetes client
func GetClient() (*kubernetes.Clientset, error) {
	if Client == nil {
		return nil, fmt.Errorf("Kubernetes client is not initialized")
	}
	return Client, nil
}

// homeDir returns the user's home directory
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // Windows
}

// getEnv retrieves environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool retrieves boolean environment variable or returns default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}
