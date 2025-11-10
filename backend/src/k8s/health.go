package k8s

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HealthChecker 處理健康檢查相關操作
type HealthChecker struct {
}

// ServiceHealth Service 健康狀態
type ServiceHealth struct {
	ServiceName string          `json:"service_name"`
	Namespace   string          `json:"namespace"`
	Healthy     bool            `json:"healthy"`
	Endpoints   []EndpointInfo  `json:"endpoints"`
	Message     string          `json:"message"`
}

// EndpointInfo Endpoint 資訊
type EndpointInfo struct {
	Address string `json:"address"`
	Ready   bool   `json:"ready"`
	PodName string `json:"pod_name,omitempty"`
}

// NewHealthChecker 建立新的健康檢查器
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{}
}

// Check 檢查 Kubernetes 叢集連線狀態
func (h *HealthChecker) Check() error {
	if IsMockMode() {
		log.Println("🔧 [MOCK] Kubernetes health check - OK")
		return nil
	}

	if Client == nil {
		return fmt.Errorf("Kubernetes client is not initialized")
	}

	ctx := context.Background()
	_, err := Client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return fmt.Errorf("failed to connect to Kubernetes cluster: %w", err)
	}

	return nil
}

// CheckServiceHealth 檢查 Service 健康狀態
func (h *HealthChecker) CheckServiceHealth(namespace, serviceName string) (*ServiceHealth, error) {
	if IsMockMode() {
		return h.checkServiceHealthMock(namespace, serviceName)
	}

	if Client == nil {
		return nil, fmt.Errorf("Kubernetes client is not initialized")
	}

	ctx := context.Background()

	// 檢查 Service 是否存在
	_, err := Client.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return &ServiceHealth{
			ServiceName: serviceName,
			Namespace:   namespace,
			Healthy:     false,
			Message:     fmt.Sprintf("Service not found: %v", err),
		}, nil
	}

	// 取得 Endpoints
	endpoints, err := h.GetEndpoints(namespace, serviceName)
	if err != nil {
		return &ServiceHealth{
			ServiceName: serviceName,
			Namespace:   namespace,
			Healthy:     false,
			Endpoints:   []EndpointInfo{},
			Message:     fmt.Sprintf("Failed to get endpoints: %v", err),
		}, nil
	}

	// 檢查是否有至少一個 Ready 的 endpoint
	hasReadyEndpoint := false
	for _, ep := range endpoints {
		if ep.Ready {
			hasReadyEndpoint = true
			break
		}
	}

	health := &ServiceHealth{
		ServiceName: serviceName,
		Namespace:   namespace,
		Healthy:     hasReadyEndpoint,
		Endpoints:   endpoints,
	}

	if hasReadyEndpoint {
		health.Message = fmt.Sprintf("Service is healthy with %d ready endpoint(s)", len(endpoints))
	} else if len(endpoints) > 0 {
		health.Message = "Service has endpoints but none are ready"
	} else {
		health.Message = "Service has no endpoints"
	}

	return health, nil
}

// GetEndpoints 取得 Service 的 Endpoints
func (h *HealthChecker) GetEndpoints(namespace, serviceName string) ([]EndpointInfo, error) {
	if IsMockMode() {
		return h.getEndpointsMock(namespace, serviceName)
	}

	if Client == nil {
		return nil, fmt.Errorf("Kubernetes client is not initialized")
	}

	ctx := context.Background()

	// 取得 Endpoints 資源
	endpoints, err := Client.CoreV1().Endpoints(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoints: %w", err)
	}

	var endpointInfos []EndpointInfo

	// 解析 endpoints
	for _, subset := range endpoints.Subsets {
		// Ready addresses
		for _, addr := range subset.Addresses {
			podName := ""
			if addr.TargetRef != nil && addr.TargetRef.Kind == "Pod" {
				podName = addr.TargetRef.Name
			}
			endpointInfos = append(endpointInfos, EndpointInfo{
				Address: addr.IP,
				Ready:   true,
				PodName: podName,
			})
		}

		// Not ready addresses
		for _, addr := range subset.NotReadyAddresses {
			podName := ""
			if addr.TargetRef != nil && addr.TargetRef.Kind == "Pod" {
				podName = addr.TargetRef.Name
			}
			endpointInfos = append(endpointInfos, EndpointInfo{
				Address: addr.IP,
				Ready:   false,
				PodName: podName,
			})
		}
	}

	return endpointInfos, nil
}

// CheckPodHealth 檢查 Pod 健康狀態
func (h *HealthChecker) CheckPodHealth(namespace, podName string) (bool, string, error) {
	if IsMockMode() {
		return h.checkPodHealthMock(namespace, podName)
	}

	if Client == nil {
		return false, "", fmt.Errorf("Kubernetes client is not initialized")
	}

	ctx := context.Background()

	pod, err := Client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return false, "", fmt.Errorf("failed to get pod: %w", err)
	}

	// 檢查 Pod 狀態
	phase := pod.Status.Phase
	if phase != corev1.PodRunning {
		return false, fmt.Sprintf("Pod is in %s phase", phase), nil
	}

	// 檢查所有容器是否 Ready
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			if condition.Status == corev1.ConditionTrue {
				return true, "Pod is running and ready", nil
			}
			return false, fmt.Sprintf("Pod is running but not ready: %s", condition.Message), nil
		}
	}

	return false, "Pod ready condition not found", nil
}

// CheckNamespaceHealth 檢查命名空間中所有 Service 的健康狀態
func (h *HealthChecker) CheckNamespaceHealth(namespace string) (map[string]*ServiceHealth, error) {
	if IsMockMode() {
		return h.checkNamespaceHealthMock(namespace)
	}

	if Client == nil {
		return nil, fmt.Errorf("Kubernetes client is not initialized")
	}

	ctx := context.Background()

	// 列出命名空間中的所有 Service
	serviceList, err := Client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	healthMap := make(map[string]*ServiceHealth)

	// 檢查每個 Service 的健康狀態
	for _, svc := range serviceList.Items {
		health, err := h.CheckServiceHealth(namespace, svc.Name)
		if err != nil {
			log.Printf("⚠️  Failed to check health for service %s/%s: %v", namespace, svc.Name, err)
			continue
		}
		healthMap[svc.Name] = health
	}

	return healthMap, nil
}

// Mock 模式實作
func (h *HealthChecker) checkServiceHealthMock(namespace, serviceName string) (*ServiceHealth, error) {
	log.Printf("🔧 [MOCK] Checking Service health: %s/%s", namespace, serviceName)

	return &ServiceHealth{
		ServiceName: serviceName,
		Namespace:   namespace,
		Healthy:     true,
		Endpoints: []EndpointInfo{
			{Address: "10.244.0.10", Ready: true, PodName: "mock-pod-1"},
			{Address: "10.244.0.11", Ready: true, PodName: "mock-pod-2"},
		},
		Message: "Service is healthy with 2 ready endpoint(s)",
	}, nil
}

func (h *HealthChecker) getEndpointsMock(namespace, serviceName string) ([]EndpointInfo, error) {
	log.Printf("🔧 [MOCK] Getting Endpoints: %s/%s", namespace, serviceName)

	return []EndpointInfo{
		{Address: "10.244.0.10", Ready: true, PodName: "mock-pod-1"},
		{Address: "10.244.0.11", Ready: true, PodName: "mock-pod-2"},
	}, nil
}

func (h *HealthChecker) checkPodHealthMock(namespace, podName string) (bool, string, error) {
	log.Printf("🔧 [MOCK] Checking Pod health: %s/%s", namespace, podName)
	return true, "Pod is running and ready", nil
}

func (h *HealthChecker) checkNamespaceHealthMock(namespace string) (map[string]*ServiceHealth, error) {
	log.Printf("🔧 [MOCK] Checking namespace health: %s", namespace)

	return map[string]*ServiceHealth{
		"mock-service-1": {
			ServiceName: "mock-service-1",
			Namespace:   namespace,
			Healthy:     true,
			Endpoints: []EndpointInfo{
				{Address: "10.244.0.10", Ready: true, PodName: "mock-pod-1"},
			},
			Message: "Service is healthy with 1 ready endpoint(s)",
		},
	}, nil
}
