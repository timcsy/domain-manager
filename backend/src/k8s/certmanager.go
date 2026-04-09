package k8s

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CertManagerHelper handles cert-manager related K8s operations
type CertManagerHelper struct{}

// NewCertManagerHelper creates a new CertManagerHelper
func NewCertManagerHelper() *CertManagerHelper {
	return &CertManagerHelper{}
}

// CreateOrUpdateCloudflareSecret creates or updates the Cloudflare API token Secret
func (h *CertManagerHelper) CreateOrUpdateCloudflareSecret(namespace, token string) error {
	if IsMockMode() {
		log.Printf("[Mock] Would create/update Cloudflare Secret in namespace %s", namespace)
		return nil
	}

	if Client == nil {
		return fmt.Errorf("Kubernetes client is not initialized")
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cloudflare-api-token",
			Namespace: namespace,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"api-token": token,
		},
	}

	ctx := context.Background()

	// Try to get existing secret
	existing, err := Client.CoreV1().Secrets(namespace).Get(ctx, "cloudflare-api-token", metav1.GetOptions{})
	if err != nil {
		// Create new
		_, err = Client.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create Cloudflare Secret: %w", err)
		}
		log.Printf("✅ Created Cloudflare API token Secret in namespace %s", namespace)
		return nil
	}

	// Update existing
	existing.StringData = secret.StringData
	_, err = Client.CoreV1().Secrets(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update Cloudflare Secret: %w", err)
	}
	log.Printf("✅ Updated Cloudflare API token Secret in namespace %s", namespace)
	return nil
}

// DeleteCloudflareSecret deletes the Cloudflare API token Secret
func (h *CertManagerHelper) DeleteCloudflareSecret(namespace string) error {
	if IsMockMode() {
		log.Printf("[Mock] Would delete Cloudflare Secret in namespace %s", namespace)
		return nil
	}

	if Client == nil {
		return fmt.Errorf("Kubernetes client is not initialized")
	}

	ctx := context.Background()
	err := Client.CoreV1().Secrets(namespace).Delete(ctx, "cloudflare-api-token", metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete Cloudflare Secret: %w", err)
	}
	log.Printf("✅ Deleted Cloudflare API token Secret from namespace %s", namespace)
	return nil
}

// ClusterIssuerConfig holds configuration for creating a ClusterIssuer
type ClusterIssuerConfig struct {
	Name               string
	Email              string
	ACMEServer         string
	IngressClass       string
	CloudflareEnabled  bool
	CloudflareSecretNS string
}

// CreateOrUpdateClusterIssuer creates or updates a ClusterIssuer with HTTP-01 and optionally DNS-01 solvers
func (h *CertManagerHelper) CreateOrUpdateClusterIssuer(cfg *ClusterIssuerConfig) error {
	if IsMockMode() {
		solverType := "HTTP-01 only"
		if cfg.CloudflareEnabled {
			solverType = "HTTP-01 + DNS-01 (Cloudflare)"
		}
		log.Printf("[Mock] Would create/update ClusterIssuer %s with %s", cfg.Name, solverType)
		return nil
	}

	if DynamicClient == nil {
		return fmt.Errorf("Kubernetes dynamic client is not initialized")
	}

	// Build solvers list
	solvers := []interface{}{}

	if cfg.CloudflareEnabled {
		// DNS-01 solver for wildcard domains
		dns01Solver := map[string]interface{}{
			"dns01": map[string]interface{}{
				"cloudflare": map[string]interface{}{
					"apiTokenSecretRef": map[string]interface{}{
						"name": "cloudflare-api-token",
						"key":  "api-token",
					},
				},
			},
		}
		solvers = append(solvers, dns01Solver)
	}

	// HTTP-01 solver as fallback
	http01Solver := map[string]interface{}{
		"http01": map[string]interface{}{
			"ingress": map[string]interface{}{
				"class": cfg.IngressClass,
			},
		},
	}
	solvers = append(solvers, http01Solver)

	issuer := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "ClusterIssuer",
			"metadata": map[string]interface{}{
				"name": cfg.Name,
			},
			"spec": map[string]interface{}{
				"acme": map[string]interface{}{
					"server": cfg.ACMEServer,
					"email":  cfg.Email,
					"privateKeySecretRef": map[string]interface{}{
						"name": cfg.Name,
					},
					"solvers": solvers,
				},
			},
		},
	}

	gvr := schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1",
		Resource: "clusterissuers",
	}

	ctx := context.Background()

	// Try to get existing
	_, err := DynamicClient.Resource(gvr).Get(ctx, cfg.Name, metav1.GetOptions{})
	if err != nil {
		// Create new
		_, err = DynamicClient.Resource(gvr).Create(ctx, issuer, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create ClusterIssuer: %w", err)
		}
		log.Printf("✅ Created ClusterIssuer %s", cfg.Name)
		return nil
	}

	// Update existing
	_, err = DynamicClient.Resource(gvr).Update(ctx, issuer, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ClusterIssuer: %w", err)
	}
	log.Printf("✅ Updated ClusterIssuer %s", cfg.Name)
	return nil
}

// GetClusterIssuerStatus checks if a ClusterIssuer exists and is ready
func (h *CertManagerHelper) GetClusterIssuerStatus(name string) (bool, error) {
	if IsMockMode() {
		return true, nil
	}

	if DynamicClient == nil {
		return false, fmt.Errorf("Kubernetes dynamic client is not initialized")
	}

	gvr := schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1",
		Resource: "clusterissuers",
	}

	ctx := context.Background()
	issuer, err := DynamicClient.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return false, nil
	}

	// Check conditions for Ready status
	conditions, found, _ := unstructured.NestedSlice(issuer.Object, "status", "conditions")
	if !found {
		return false, nil
	}

	for _, c := range conditions {
		cond, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		if cond["type"] == "Ready" && cond["status"] == "True" {
			return true, nil
		}
	}

	return false, nil
}
