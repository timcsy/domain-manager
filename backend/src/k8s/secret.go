package k8s

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecretManager 處理 Secret 資源操作
type SecretManager struct {
}

// NewSecretManager 建立新的 Secret 管理器
func NewSecretManager() *SecretManager {
	return &SecretManager{}
}

// CreateTLSSecret 建立 TLS Secret
func (m *SecretManager) CreateTLSSecret(namespace, name, certPEM, keyPEM string) (*corev1.Secret, error) {
	if IsMockMode() {
		return m.createTLSSecretMock(namespace, name, certPEM, keyPEM)
	}

	if Client == nil {
		return nil, fmt.Errorf("Kubernetes client is not initialized")
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       []byte(certPEM),
			corev1.TLSPrivateKeyKey: []byte(keyPEM),
		},
	}

	ctx := context.Background()
	result, err := Client.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS secret: %w", err)
	}

	log.Printf("✅ Created TLS Secret: %s/%s", namespace, name)
	return result, nil
}

// UpdateSecret 更新 Secret
func (m *SecretManager) UpdateSecret(namespace, name, certPEM, keyPEM string) (*corev1.Secret, error) {
	if IsMockMode() {
		return m.updateSecretMock(namespace, name, certPEM, keyPEM)
	}

	if Client == nil {
		return nil, fmt.Errorf("Kubernetes client is not initialized")
	}

	ctx := context.Background()

	// 先取得現有的 Secret
	existingSecret, err := Client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get existing secret: %w", err)
	}

	// 更新 Secret 資料
	existingSecret.Data = map[string][]byte{
		corev1.TLSCertKey:       []byte(certPEM),
		corev1.TLSPrivateKeyKey: []byte(keyPEM),
	}

	result, err := Client.CoreV1().Secrets(namespace).Update(ctx, existingSecret, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update secret: %w", err)
	}

	log.Printf("✅ Updated Secret: %s/%s", namespace, name)
	return result, nil
}

// DeleteSecret 刪除 Secret
func (m *SecretManager) DeleteSecret(namespace, name string) error {
	if IsMockMode() {
		return m.deleteSecretMock(namespace, name)
	}

	if Client == nil {
		return fmt.Errorf("Kubernetes client is not initialized")
	}

	ctx := context.Background()
	err := Client.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	log.Printf("✅ Deleted Secret: %s/%s", namespace, name)
	return nil
}

// GetSecret 取得 Secret
func (m *SecretManager) GetSecret(namespace, name string) (*corev1.Secret, error) {
	if IsMockMode() {
		return m.getSecretMock(namespace, name)
	}

	if Client == nil {
		return nil, fmt.Errorf("Kubernetes client is not initialized")
	}

	ctx := context.Background()
	secret, err := Client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	return secret, nil
}

// SecretExists 檢查 Secret 是否存在
func (m *SecretManager) SecretExists(namespace, name string) (bool, error) {
	if IsMockMode() {
		return m.secretExistsMock(namespace, name)
	}

	if Client == nil {
		return false, fmt.Errorf("Kubernetes client is not initialized")
	}

	ctx := context.Background()
	_, err := Client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		// 如果是 NotFound 錯誤，回傳 false
		if err.Error() == "not found" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check secret existence: %w", err)
	}

	return true, nil
}

// Mock 模式實作
func (m *SecretManager) createTLSSecretMock(namespace, name, certPEM, keyPEM string) (*corev1.Secret, error) {
	log.Printf("🔧 [MOCK] Creating TLS Secret: %s/%s (cert: %d bytes, key: %d bytes)",
		namespace, name, len(certPEM), len(keyPEM))

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       []byte(certPEM),
			corev1.TLSPrivateKeyKey: []byte(keyPEM),
		},
	}, nil
}

func (m *SecretManager) updateSecretMock(namespace, name, certPEM, keyPEM string) (*corev1.Secret, error) {
	log.Printf("🔧 [MOCK] Updating Secret: %s/%s", namespace, name)
	return m.createTLSSecretMock(namespace, name, certPEM, keyPEM)
}

func (m *SecretManager) deleteSecretMock(namespace, name string) error {
	log.Printf("🔧 [MOCK] Deleting Secret: %s/%s", namespace, name)
	return nil
}

func (m *SecretManager) getSecretMock(namespace, name string) (*corev1.Secret, error) {
	log.Printf("🔧 [MOCK] Getting Secret: %s/%s", namespace, name)
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       []byte("mock-cert-pem"),
			corev1.TLSPrivateKeyKey: []byte("mock-key-pem"),
		},
	}, nil
}

func (m *SecretManager) secretExistsMock(namespace, name string) (bool, error) {
	log.Printf("🔧 [MOCK] Checking if Secret exists: %s/%s", namespace, name)
	return true, nil
}
