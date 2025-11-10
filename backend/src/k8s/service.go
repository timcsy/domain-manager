package k8s

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceManager 處理 Service 資源操作
type ServiceManager struct {
}

// ServiceInfo Service 資訊
type ServiceInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Type      string            `json:"type"`
	ClusterIP string            `json:"cluster_ip"`
	Ports     []ServicePort     `json:"ports"`
	Selector  map[string]string `json:"selector"`
}

// ServicePort Service 埠號資訊
type ServicePort struct {
	Name       string `json:"name"`
	Port       int32  `json:"port"`
	TargetPort string `json:"target_port"`
	Protocol   string `json:"protocol"`
}

// NewServiceManager 建立新的 Service 管理器
func NewServiceManager() *ServiceManager {
	return &ServiceManager{}
}

// ListServices 列出命名空間中的所有 Service
func (m *ServiceManager) ListServices(namespace string) ([]ServiceInfo, error) {
	if IsMockMode() {
		return m.listServicesMock(namespace)
	}

	if Client == nil {
		return nil, fmt.Errorf("Kubernetes client is not initialized")
	}

	ctx := context.Background()
	serviceList, err := Client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	var services []ServiceInfo
	for _, svc := range serviceList.Items {
		services = append(services, m.convertServiceToInfo(&svc))
	}

	return services, nil
}

// GetService 取得特定 Service
func (m *ServiceManager) GetService(namespace, name string) (*ServiceInfo, error) {
	if IsMockMode() {
		return m.getServiceMock(namespace, name)
	}

	if Client == nil {
		return nil, fmt.Errorf("Kubernetes client is not initialized")
	}

	ctx := context.Background()
	svc, err := Client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	info := m.convertServiceToInfo(svc)
	return &info, nil
}

// ValidateService 驗證 Service 是否存在且有效
func (m *ServiceManager) ValidateService(namespace, name string, port int) error {
	if IsMockMode() {
		return m.validateServiceMock(namespace, name, port)
	}

	if Client == nil {
		return fmt.Errorf("Kubernetes client is not initialized")
	}

	ctx := context.Background()
	svc, err := Client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("service not found: %w", err)
	}

	// 驗證埠號是否存在
	portFound := false
	for _, p := range svc.Spec.Ports {
		if int(p.Port) == port {
			portFound = true
			break
		}
	}

	if !portFound {
		return fmt.Errorf("port %d not found in service %s/%s", port, namespace, name)
	}

	return nil
}

// ServiceExists 檢查 Service 是否存在
func (m *ServiceManager) ServiceExists(namespace, name string) (bool, error) {
	if IsMockMode() {
		return m.serviceExistsMock(namespace, name)
	}

	if Client == nil {
		return false, fmt.Errorf("Kubernetes client is not initialized")
	}

	ctx := context.Background()
	_, err := Client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		// 如果是 NotFound 錯誤，回傳 false
		if err.Error() == "not found" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check service existence: %w", err)
	}

	return true, nil
}

// convertServiceToInfo 將 K8s Service 轉換為 ServiceInfo
func (m *ServiceManager) convertServiceToInfo(svc *corev1.Service) ServiceInfo {
	var ports []ServicePort
	for _, p := range svc.Spec.Ports {
		ports = append(ports, ServicePort{
			Name:       p.Name,
			Port:       p.Port,
			TargetPort: p.TargetPort.String(),
			Protocol:   string(p.Protocol),
		})
	}

	return ServiceInfo{
		Name:      svc.Name,
		Namespace: svc.Namespace,
		Type:      string(svc.Spec.Type),
		ClusterIP: svc.Spec.ClusterIP,
		Ports:     ports,
		Selector:  svc.Spec.Selector,
	}
}

// Mock 模式實作
func (m *ServiceManager) listServicesMock(namespace string) ([]ServiceInfo, error) {
	log.Printf("🔧 [MOCK] Listing Services in namespace: %s", namespace)

	return []ServiceInfo{
		{
			Name:      "mock-service-1",
			Namespace: namespace,
			Type:      "ClusterIP",
			ClusterIP: "10.96.0.1",
			Ports: []ServicePort{
				{Name: "http", Port: 80, TargetPort: "8080", Protocol: "TCP"},
			},
			Selector: map[string]string{"app": "mock-app"},
		},
		{
			Name:      "mock-service-2",
			Namespace: namespace,
			Type:      "ClusterIP",
			ClusterIP: "10.96.0.2",
			Ports: []ServicePort{
				{Name: "http", Port: 80, TargetPort: "8080", Protocol: "TCP"},
				{Name: "https", Port: 443, TargetPort: "8443", Protocol: "TCP"},
			},
			Selector: map[string]string{"app": "mock-app-2"},
		},
	}, nil
}

func (m *ServiceManager) getServiceMock(namespace, name string) (*ServiceInfo, error) {
	log.Printf("🔧 [MOCK] Getting Service: %s/%s", namespace, name)

	return &ServiceInfo{
		Name:      name,
		Namespace: namespace,
		Type:      "ClusterIP",
		ClusterIP: "10.96.0.100",
		Ports: []ServicePort{
			{Name: "http", Port: 80, TargetPort: "8080", Protocol: "TCP"},
		},
		Selector: map[string]string{"app": name},
	}, nil
}

func (m *ServiceManager) validateServiceMock(namespace, name string, port int) error {
	log.Printf("🔧 [MOCK] Validating Service: %s/%s:%d", namespace, name, port)
	// Mock 模式下總是驗證通過
	return nil
}

func (m *ServiceManager) serviceExistsMock(namespace, name string) (bool, error) {
	log.Printf("🔧 [MOCK] Checking if Service exists: %s/%s", namespace, name)
	return true, nil
}
