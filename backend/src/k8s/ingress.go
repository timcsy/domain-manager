package k8s

import (
	"context"
	"fmt"
	"log"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IngressManager 處理 Ingress 資源操作
type IngressManager struct {
}

// IngressConfig Ingress 配置
type IngressConfig struct {
	Name             string
	Namespace        string
	Host             string
	ServiceName      string
	ServicePort      int
	TLSSecretName    string
	IngressClassName *string
	Annotations      map[string]string
}

// NewIngressManager 建立新的 Ingress 管理器
func NewIngressManager() *IngressManager {
	return &IngressManager{}
}

// CreateIngress 建立 Ingress 資源
func (m *IngressManager) CreateIngress(cfg *IngressConfig) (*networkingv1.Ingress, error) {
	if IsMockMode() {
		return m.createIngressMock(cfg)
	}

	if Client == nil {
		return nil, fmt.Errorf("Kubernetes client is not initialized")
	}

	pathType := networkingv1.PathTypePrefix
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cfg.Name,
			Namespace:   cfg.Namespace,
			Annotations: cfg.Annotations,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: cfg.IngressClassName,
			Rules: []networkingv1.IngressRule{
				{
					Host: cfg.Host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: cfg.ServiceName,
											Port: networkingv1.ServiceBackendPort{
												Number: int32(cfg.ServicePort),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// 如果有 TLS 配置，添加 TLS 設定
	if cfg.TLSSecretName != "" {
		ingress.Spec.TLS = []networkingv1.IngressTLS{
			{
				Hosts:      []string{cfg.Host},
				SecretName: cfg.TLSSecretName,
			},
		}
	}

	ctx := context.Background()
	result, err := Client.NetworkingV1().Ingresses(cfg.Namespace).Create(ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create ingress: %w", err)
	}

	log.Printf("✅ Created Ingress: %s/%s for host %s", cfg.Namespace, cfg.Name, cfg.Host)
	return result, nil
}

// UpdateIngress 更新 Ingress 資源
func (m *IngressManager) UpdateIngress(cfg *IngressConfig) (*networkingv1.Ingress, error) {
	if IsMockMode() {
		return m.updateIngressMock(cfg)
	}

	if Client == nil {
		return nil, fmt.Errorf("Kubernetes client is not initialized")
	}

	ctx := context.Background()

	// 先取得現有的 Ingress
	existingIngress, err := Client.NetworkingV1().Ingresses(cfg.Namespace).Get(ctx, cfg.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get existing ingress: %w", err)
	}

	// 更新配置
	pathType := networkingv1.PathTypePrefix
	existingIngress.Spec.IngressClassName = cfg.IngressClassName
	existingIngress.Spec.Rules = []networkingv1.IngressRule{
		{
			Host: cfg.Host,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: []networkingv1.HTTPIngressPath{
						{
							Path:     "/",
							PathType: &pathType,
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: cfg.ServiceName,
									Port: networkingv1.ServiceBackendPort{
										Number: int32(cfg.ServicePort),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// 更新 TLS 配置
	if cfg.TLSSecretName != "" {
		existingIngress.Spec.TLS = []networkingv1.IngressTLS{
			{
				Hosts:      []string{cfg.Host},
				SecretName: cfg.TLSSecretName,
			},
		}
	} else {
		existingIngress.Spec.TLS = nil
	}

	// 更新 annotations
	if cfg.Annotations != nil {
		existingIngress.Annotations = cfg.Annotations
	}

	result, err := Client.NetworkingV1().Ingresses(cfg.Namespace).Update(ctx, existingIngress, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update ingress: %w", err)
	}

	log.Printf("✅ Updated Ingress: %s/%s", cfg.Namespace, cfg.Name)
	return result, nil
}

// DeleteIngress 刪除 Ingress 資源
func (m *IngressManager) DeleteIngress(namespace, name string) error {
	if IsMockMode() {
		return m.deleteIngressMock(namespace, name)
	}

	if Client == nil {
		return fmt.Errorf("Kubernetes client is not initialized")
	}

	ctx := context.Background()
	err := Client.NetworkingV1().Ingresses(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete ingress: %w", err)
	}

	log.Printf("✅ Deleted Ingress: %s/%s", namespace, name)
	return nil
}

// GetIngress 取得 Ingress 資源
func (m *IngressManager) GetIngress(namespace, name string) (*networkingv1.Ingress, error) {
	if IsMockMode() {
		return m.getIngressMock(namespace, name)
	}

	if Client == nil {
		return nil, fmt.Errorf("Kubernetes client is not initialized")
	}

	ctx := context.Background()
	ingress, err := Client.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ingress: %w", err)
	}

	return ingress, nil
}

// ListIngresses 列出命名空間中的所有 Ingress
func (m *IngressManager) ListIngresses(namespace string) (*networkingv1.IngressList, error) {
	if IsMockMode() {
		return m.listIngressesMock(namespace)
	}

	if Client == nil {
		return nil, fmt.Errorf("Kubernetes client is not initialized")
	}

	ctx := context.Background()
	ingresses, err := Client.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list ingresses: %w", err)
	}

	return ingresses, nil
}

// Mock 模式實作
func (m *IngressManager) createIngressMock(cfg *IngressConfig) (*networkingv1.Ingress, error) {
	log.Printf("🔧 [MOCK] Creating Ingress: %s/%s for host %s → %s:%d",
		cfg.Namespace, cfg.Name, cfg.Host, cfg.ServiceName, cfg.ServicePort)

	pathType := networkingv1.PathTypePrefix
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: cfg.Host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: cfg.ServiceName,
											Port: networkingv1.ServiceBackendPort{
												Number: int32(cfg.ServicePort),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if cfg.TLSSecretName != "" {
		ingress.Spec.TLS = []networkingv1.IngressTLS{
			{
				Hosts:      []string{cfg.Host},
				SecretName: cfg.TLSSecretName,
			},
		}
	}

	return ingress, nil
}

func (m *IngressManager) updateIngressMock(cfg *IngressConfig) (*networkingv1.Ingress, error) {
	log.Printf("🔧 [MOCK] Updating Ingress: %s/%s", cfg.Namespace, cfg.Name)
	return m.createIngressMock(cfg)
}

func (m *IngressManager) deleteIngressMock(namespace, name string) error {
	log.Printf("🔧 [MOCK] Deleting Ingress: %s/%s", namespace, name)
	return nil
}

func (m *IngressManager) getIngressMock(namespace, name string) (*networkingv1.Ingress, error) {
	log.Printf("🔧 [MOCK] Getting Ingress: %s/%s", namespace, name)
	pathType := networkingv1.PathTypePrefix
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: "mock.example.com",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "mock-service",
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

func (m *IngressManager) listIngressesMock(namespace string) (*networkingv1.IngressList, error) {
	log.Printf("🔧 [MOCK] Listing Ingresses in namespace: %s", namespace)
	return &networkingv1.IngressList{
		Items: []networkingv1.Ingress{},
	}, nil
}
