package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"sync"
)

// KeyManager handles RSA key operations and caching
type KeyManager struct {
	privateKey     *rsa.PrivateKey
	privateKeyPath string
	publicKeyPath  string
	mu             sync.RWMutex
}

// NewKeyManager creates a new KeyManager instance
func NewKeyManager(privateKeyPath, publicKeyPath string) *KeyManager {
	return &KeyManager{
		privateKeyPath: privateKeyPath,
		publicKeyPath:  publicKeyPath,
	}
}

// EnsureKeys checks if keys exist and generates them if they don't
func (km *KeyManager) EnsureKeys() error {
	// First check without lock
	if km.privateKey != nil {
		return nil
	}

	// Check if private key exists before acquiring lock
	keyExists := false
	if _, err := os.Stat(km.privateKeyPath); err == nil {
		keyExists = true
	}

	// Generate keys if needed - do this outside the lock
	if !keyExists {
		if err := km.generateRSAKeys(); err != nil {
			return fmt.Errorf("failed to generate RSA keys: %w", err)
		}
	}

	// Now acquire lock only for updating the in-memory key
	km.mu.Lock()
	defer km.mu.Unlock()

	// Double-check if key was loaded by another goroutine
	if km.privateKey != nil {
		return nil
	}

	// Load the private key
	privateKey, err := km.LoadPrivateKey()
	if err != nil {
		return fmt.Errorf("failed to load private key: %w", err)
	}

	km.privateKey = privateKey
	return nil
}

// GetPublicKeyPEM returns the PEM encoded public key
func (km *KeyManager) GetPublicKeyPEM() ([]byte, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if km.privateKey == nil {
		return nil, errors.New("private key not loaded")
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&km.privateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return publicKeyPEM, nil
}

// GetPrivateKey returns the RSA private key
func (km *KeyManager) GetPrivateKey() *rsa.PrivateKey {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.privateKey
}

func (km *KeyManager) generateRSAKeys() error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Save private key
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	if err := os.WriteFile(km.privateKeyPath, privateKeyPEM, 0600); err != nil {
		return err
	}

	// Save public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return os.WriteFile(km.publicKeyPath, publicKeyPEM, 0644)
}

// LoadPrivateKey loads the private key from file
func (km *KeyManager) LoadPrivateKey() (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(km.privateKeyPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
