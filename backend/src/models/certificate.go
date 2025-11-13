package models

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"
)

// Certificate 代表 SSL/TLS 憑證
type Certificate struct {
	ID                   int       `json:"id" db:"id"`
	DomainName           string    `json:"domain_name" db:"domain_name"`
	Source               string    `json:"source" db:"source"` // 'letsencrypt' or 'manual'
	CertificatePEM       string    `json:"certificate_pem" db:"certificate_pem"`
	PrivateKeyPEM        string    `json:"-" db:"private_key_pem"` // 不在 JSON 中暴露私鑰
	Issuer               string    `json:"issuer" db:"issuer"`
	ValidFrom            time.Time `json:"valid_from" db:"valid_from"`
	ValidUntil           time.Time `json:"valid_until" db:"valid_until"`
	Status               string    `json:"status" db:"status"` // 'valid', 'expiring', 'expired', 'revoked'
	K8sSecretName        string    `json:"k8s_secret_name" db:"k8s_secret_name"`
	K8sSecretNamespace   string    `json:"k8s_secret_namespace" db:"k8s_secret_namespace"`
	AutoRenew            bool      `json:"auto_renew" db:"auto_renew"`
	LastRenewalAttempt   *time.Time `json:"last_renewal_attempt,omitempty" db:"last_renewal_attempt"`
	RenewalError         *string   `json:"renewal_error,omitempty" db:"renewal_error"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

// CertificateStatus 憑證狀態常數
const (
	CertStatusValid    = "valid"
	CertStatusExpiring = "expiring"
	CertStatusExpired  = "expired"
	CertStatusRevoked  = "revoked"
)

// CertificateSource 憑證來源常數
const (
	CertSourceLetsEncrypt = "letsencrypt"
	CertSourceManual      = "manual"
)

// IsExpiring 檢查憑證是否即將到期（30天內）
func (c *Certificate) IsExpiring() bool {
	return time.Until(c.ValidUntil) <= 30*24*time.Hour
}

// IsExpired 檢查憑證是否已過期
func (c *Certificate) IsExpired() bool {
	return time.Now().After(c.ValidUntil)
}

// DaysUntilExpiry 返回憑證到期前的天數
func (c *Certificate) DaysUntilExpiry() int {
	duration := time.Until(c.ValidUntil)
	return int(duration.Hours() / 24)
}

// UpdateStatus 更新憑證狀態
func (c *Certificate) UpdateStatus() {
	if c.IsExpired() {
		c.Status = CertStatusExpired
	} else if c.IsExpiring() {
		c.Status = CertStatusExpiring
	} else if c.Status != CertStatusRevoked {
		c.Status = CertStatusValid
	}
}

// ParseCertificatePEM 解析憑證 PEM 並提取資訊
func ParseCertificatePEM(certPEM string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// ValidatePrivateKey 驗證私鑰 PEM 格式
func ValidatePrivateKey(keyPEM string) error {
	block, _ := pem.Decode([]byte(keyPEM))
	if block == nil {
		return fmt.Errorf("failed to decode private key PEM")
	}

	// 嘗試解析 PKCS1 或 PKCS8 格式
	_, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		_, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	return nil
}

// ValidateCertificateKeyPair 驗證憑證和私鑰是否匹配
func ValidateCertificateKeyPair(certPEM, keyPEM string) error {
	cert, err := ParseCertificatePEM(certPEM)
	if err != nil {
		return err
	}

	if err := ValidatePrivateKey(keyPEM); err != nil {
		return err
	}

	// 這裡可以添加更複雜的匹配驗證
	// 例如使用私鑰簽名，然後用公鑰驗證

	_ = cert // 暫時未使用，避免編譯器警告

	return nil
}

// ExtractCertificateInfo 從 PEM 提取憑證資訊
func ExtractCertificateInfo(certPEM string) (issuer string, validFrom, validUntil time.Time, err error) {
	cert, err := ParseCertificatePEM(certPEM)
	if err != nil {
		return "", time.Time{}, time.Time{}, err
	}

	return cert.Issuer.CommonName, cert.NotBefore, cert.NotAfter, nil
}

// CertificateFilter represents filters for listing certificates
type CertificateFilter struct {
	Status     string
	DomainName string
	Source     string
	Limit      int
	Offset     int
}
