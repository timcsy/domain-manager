package certificate

import (
	"os"
	"strings"
	"testing"
)

func TestEncryptor(t *testing.T) {
	// Set test encryption key
	os.Setenv("ENCRYPTION_KEY", "test-encryption-key-for-unit-tests")
	defer os.Unsetenv("ENCRYPTION_KEY")

	encryptor, err := NewEncryptor()
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	t.Run("Basic encryption and decryption", func(t *testing.T) {
		plaintext := "Hello, World!"

		// Encrypt
		ciphertext, err := encryptor.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("Encryption failed: %v", err)
		}

		// Verify ciphertext is different from plaintext
		if ciphertext == plaintext {
			t.Error("Ciphertext should be different from plaintext")
		}

		// Decrypt
		decrypted, err := encryptor.Decrypt(ciphertext)
		if err != nil {
			t.Fatalf("Decryption failed: %v", err)
		}

		// Verify decrypted matches original
		if decrypted != plaintext {
			t.Errorf("Decrypted text doesn't match original. Got %q, want %q", decrypted, plaintext)
		}
	})

	t.Run("Private key encryption", func(t *testing.T) {
		privateKeyPEM := `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC1234567890ABC
DEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789ABCDEF
-----END PRIVATE KEY-----`

		// Encrypt private key
		encrypted, err := encryptor.EncryptPrivateKey(privateKeyPEM)
		if err != nil {
			t.Fatalf("Private key encryption failed: %v", err)
		}

		// Verify encrypted is different from original
		if encrypted == privateKeyPEM {
			t.Error("Encrypted private key should be different from original")
		}

		// Decrypt private key
		decrypted, err := encryptor.DecryptPrivateKey(encrypted)
		if err != nil {
			t.Fatalf("Private key decryption failed: %v", err)
		}

		// Verify decrypted matches original
		if decrypted != privateKeyPEM {
			t.Error("Decrypted private key doesn't match original")
		}
	})

	t.Run("Empty string handling", func(t *testing.T) {
		// Try to encrypt empty string
		_, err := encryptor.Encrypt("")
		if err == nil {
			t.Error("Expected error when encrypting empty string")
		}

		// Try to decrypt empty string
		_, err = encryptor.Decrypt("")
		if err == nil {
			t.Error("Expected error when decrypting empty string")
		}
	})

	t.Run("Invalid ciphertext handling", func(t *testing.T) {
		// Try to decrypt invalid ciphertext
		_, err := encryptor.Decrypt("invalid-ciphertext")
		if err == nil {
			t.Error("Expected error when decrypting invalid ciphertext")
		}
	})

	t.Run("Deterministic nonce", func(t *testing.T) {
		plaintext := "Test message"

		// Encrypt twice
		encrypted1, _ := encryptor.Encrypt(plaintext)
		encrypted2, _ := encryptor.Encrypt(plaintext)

		// Verify ciphertexts are different (due to random nonce)
		if encrypted1 == encrypted2 {
			t.Error("Same plaintext should produce different ciphertexts (random nonce)")
		}

		// But both should decrypt to the same plaintext
		decrypted1, _ := encryptor.Decrypt(encrypted1)
		decrypted2, _ := encryptor.Decrypt(encrypted2)

		if decrypted1 != plaintext || decrypted2 != plaintext {
			t.Error("Both ciphertexts should decrypt to the same plaintext")
		}
	})
}

func TestNewEncryptor_MissingKey(t *testing.T) {
	// Ensure ENCRYPTION_KEY is not set
	os.Unsetenv("ENCRYPTION_KEY")

	_, err := NewEncryptor()
	if err == nil {
		t.Error("Expected error when ENCRYPTION_KEY is not set")
	}

	if !strings.Contains(err.Error(), "ENCRYPTION_KEY") {
		t.Errorf("Error should mention ENCRYPTION_KEY, got: %v", err)
	}
}
