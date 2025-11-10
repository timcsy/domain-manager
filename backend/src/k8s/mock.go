package k8s

import (
	"fmt"
	"log"
)

// MockClient is a mock Kubernetes client for local development
type MockClient struct {
	enabled bool
}

// NewMockClient creates a new mock K8s client
func NewMockClient() *MockClient {
	return &MockClient{enabled: true}
}

// InitializeMock initializes mock mode (no real K8s connection)
func InitializeMock() error {
	log.Println("⚠️  Running in MOCK mode - Kubernetes operations will be simulated")
	log.Println("   Set K8S_MOCK=false to connect to real cluster")

	// Don't actually initialize K8s client
	Client = nil

	return nil
}

// HealthMock checks if mock mode is healthy
func HealthMock() error {
	log.Println("Mock K8s Health Check: OK")
	return nil
}

// IsMockMode checks if running in mock mode
func IsMockMode() bool {
	return Client == nil
}

// MockCreateIngress simulates creating an ingress
func MockCreateIngress(domain, service, namespace string, port int) error {
	log.Printf("📝 [MOCK] Create Ingress: %s -> %s.%s:%d", domain, service, namespace, port)
	return nil
}

// MockDeleteIngress simulates deleting an ingress
func MockDeleteIngress(domain string) error {
	log.Printf("🗑️  [MOCK] Delete Ingress: %s", domain)
	return nil
}

// MockCreateSecret simulates creating a secret
func MockCreateSecret(name, namespace string, data map[string][]byte) error {
	log.Printf("🔒 [MOCK] Create Secret: %s/%s", namespace, name)
	return nil
}

// MockListServices simulates listing services
func MockListServices(namespace string) ([]MockService, error) {
	log.Printf("📋 [MOCK] List Services in namespace: %s", namespace)

	// Return some mock services
	mockServices := []MockService{
		{Name: "example-service", Namespace: "default", Port: 80},
		{Name: "api-service", Namespace: "default", Port: 8080},
		{Name: "web-service", Namespace: "default", Port: 3000},
	}

	return mockServices, nil
}

// MockService represents a mock Kubernetes service
type MockService struct {
	Name      string
	Namespace string
	Port      int
}

// ValidateConfig validates configuration in mock mode
func ValidateConfig() error {
	if IsMockMode() {
		log.Println("✅ Mock mode validated")
		return nil
	}
	return fmt.Errorf("not in mock mode")
}
