package letsencrypt

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

// User implements the ACME User interface for lego
type User struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

// GetEmail returns the user's email
func (u *User) GetEmail() string {
	return u.Email
}

// GetRegistration returns the user's registration resource
func (u *User) GetRegistration() *registration.Resource {
	return u.Registration
}

// GetPrivateKey returns the user's private key
func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

// Client wraps the ACME client with configuration
type Client struct {
	client      *lego.Client
	user        *User
	mu          sync.RWMutex
	accountPath string
	staging     bool
}

// Config holds the Let's Encrypt client configuration
type Config struct {
	Email       string // Contact email for Let's Encrypt
	AccountPath string // Path to store account information
	Staging     bool   // Use staging environment for testing
}

var (
	globalClient *Client
	clientMu     sync.RWMutex
)

// Initialize creates and configures the global Let's Encrypt client
func Initialize(cfg *Config) error {
	clientMu.Lock()
	defer clientMu.Unlock()

	if cfg.Email == "" {
		return fmt.Errorf("email is required for Let's Encrypt registration")
	}

	if cfg.AccountPath == "" {
		cfg.AccountPath = "./data/letsencrypt"
	}

	// Ensure account directory exists
	if err := os.MkdirAll(cfg.AccountPath, 0700); err != nil {
		return fmt.Errorf("failed to create account directory: %w", err)
	}

	// Load or create user account
	user, err := loadOrCreateUser(cfg.Email, cfg.AccountPath)
	if err != nil {
		return fmt.Errorf("failed to load/create user: %w", err)
	}

	// Create lego config
	legoConfig := lego.NewConfig(user)

	// Set CA directory URL based on staging flag
	if cfg.Staging {
		legoConfig.CADirURL = lego.LEDirectoryStaging
		log.Println("Using Let's Encrypt STAGING environment")
	} else {
		legoConfig.CADirURL = lego.LEDirectoryProduction
		log.Println("Using Let's Encrypt PRODUCTION environment")
	}

	// Create ACME client
	client, err := lego.NewClient(legoConfig)
	if err != nil {
		return fmt.Errorf("failed to create ACME client: %w", err)
	}

	// Register if needed
	if user.Registration == nil {
		reg, err := client.Registration.Register(registration.RegisterOptions{
			TermsOfServiceAgreed: true,
		})
		if err != nil {
			return fmt.Errorf("failed to register with ACME: %w", err)
		}
		user.Registration = reg

		// Save the registration
		if err := saveUser(user, cfg.AccountPath); err != nil {
			log.Printf("Warning: failed to save user registration: %v", err)
		}
	}

	globalClient = &Client{
		client:      client,
		user:        user,
		accountPath: cfg.AccountPath,
		staging:     cfg.Staging,
	}

	log.Printf("Let's Encrypt client initialized for: %s", cfg.Email)
	return nil
}

// GetClient returns the global Let's Encrypt client
func GetClient() (*Client, error) {
	clientMu.RLock()
	defer clientMu.RUnlock()

	if globalClient == nil {
		return nil, fmt.Errorf("Let's Encrypt client not initialized")
	}

	return globalClient, nil
}

// loadOrCreateUser loads an existing user or creates a new one
func loadOrCreateUser(email, accountPath string) (*User, error) {
	userFile := filepath.Join(accountPath, "user.json")
	keyFile := filepath.Join(accountPath, "user.key")

	// Try to load existing user
	if _, err := os.Stat(userFile); err == nil {
		return loadUser(userFile, keyFile)
	}

	// Create new user
	log.Printf("Creating new Let's Encrypt account for: %s", email)

	// Generate private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	user := &User{
		Email: email,
		key:   privateKey,
	}

	// Save the new user
	if err := saveUser(user, accountPath); err != nil {
		return nil, fmt.Errorf("failed to save new user: %w", err)
	}

	return user, nil
}

// loadUser loads user information from files
func loadUser(userFile, keyFile string) (*User, error) {
	// Load user data
	userData, err := os.ReadFile(userFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read user file: %w", err)
	}

	var user User
	if err := json.Unmarshal(userData, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	// Load private key
	keyData, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM key")
	}

	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	user.key = privateKey
	return &user, nil
}

// saveUser saves user information to files
func saveUser(user *User, accountPath string) error {
	userFile := filepath.Join(accountPath, "user.json")
	keyFile := filepath.Join(accountPath, "user.key")

	// Save user data
	userData, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	if err := os.WriteFile(userFile, userData, 0600); err != nil {
		return fmt.Errorf("failed to write user file: %w", err)
	}

	// Save private key
	keyBytes, err := x509.MarshalECPrivateKey(user.key.(*ecdsa.PrivateKey))
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	pemBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	}

	if err := os.WriteFile(keyFile, pem.EncodeToMemory(pemBlock), 0600); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	return nil
}

// IsStaging returns whether the client is using the staging environment
func (c *Client) IsStaging() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.staging
}

// GetEmail returns the user's email
func (c *Client) GetEmail() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.user.Email
}
