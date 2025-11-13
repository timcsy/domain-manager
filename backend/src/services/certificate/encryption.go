package certificate

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

// Encryptor handles certificate encryption/decryption
type Encryptor struct {
	key []byte
}

// NewEncryptor creates a new encryptor with the given encryption key
func NewEncryptor() (*Encryptor, error) {
	// Get encryption key from environment
	keyString := os.Getenv("ENCRYPTION_KEY")
	if keyString == "" {
		return nil, fmt.Errorf("ENCRYPTION_KEY environment variable is not set")
	}

	// Derive a 256-bit key using PBKDF2
	salt := []byte("domain-manager-salt") // In production, use a unique salt per deployment
	key := pbkdf2.Key([]byte(keyString), salt, 10000, 32, sha256.New)

	return &Encryptor{
		key: key,
	}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", fmt.Errorf("plaintext cannot be empty")
	}

	// Create AES cipher
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the plaintext
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-256-GCM
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", fmt.Errorf("ciphertext cannot be empty")
	}

	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]

	// Decrypt the ciphertext
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptPrivateKey encrypts a private key PEM
func (e *Encryptor) EncryptPrivateKey(privateKeyPEM string) (string, error) {
	return e.Encrypt(privateKeyPEM)
}

// DecryptPrivateKey decrypts a private key PEM
func (e *Encryptor) DecryptPrivateKey(encryptedPrivateKey string) (string, error) {
	return e.Decrypt(encryptedPrivateKey)
}
