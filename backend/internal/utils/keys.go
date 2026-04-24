package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// KeyManager handles RSA key pair loading, generation, and lifecycle.
type KeyManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	mu         sync.RWMutex
}

// NewKeyManager loads or generates RSA keys.
// Priority: 1) JWT_PRIVATE_KEY / JWT_PUBLIC_KEY env vars (for cloud deployments)
//           2) PEM files on disk (for local development)
//           3) Auto-generate new keys and save to disk
func NewKeyManager(privateKeyPath, publicKeyPath string) (*KeyManager, error) {
	km := &KeyManager{}

	// --- Cloud mode: load keys from environment variables ---
	privEnv := os.Getenv("JWT_PRIVATE_KEY")
	pubEnv := os.Getenv("JWT_PUBLIC_KEY")
	if privEnv != "" && pubEnv != "" {
		if err := km.loadKeysFromPEM([]byte(privEnv), []byte(pubEnv)); err != nil {
			return nil, fmt.Errorf("failed to load keys from env vars: %w", err)
		}
		return km, nil
	}

	// --- Local mode: load from files or generate ---
	privExists := fileExists(privateKeyPath)
	pubExists := fileExists(publicKeyPath)

	if privExists && pubExists {
		if err := km.loadKeys(privateKeyPath, publicKeyPath); err != nil {
			return nil, fmt.Errorf("failed to load existing keys: %w", err)
		}
		return km, nil
	}

	// Generate new key pair
	if err := km.generateAndSaveKeys(privateKeyPath, publicKeyPath); err != nil {
		return nil, fmt.Errorf("failed to generate keys: %w", err)
	}

	return km, nil
}

// PrivateKey returns the RSA private key (for signing tokens).
func (km *KeyManager) PrivateKey() *rsa.PrivateKey {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.privateKey
}

// PublicKey returns the RSA public key (for verifying tokens).
func (km *KeyManager) PublicKey() *rsa.PublicKey {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.publicKey
}

// loadKeysFromPEM parses RSA keys from raw PEM byte slices (used with env vars).
func (km *KeyManager) loadKeysFromPEM(privPEM, pubPEM []byte) error {
	privBlock, _ := pem.Decode(privPEM)
	if privBlock == nil {
		return fmt.Errorf("failed to decode private key PEM block from env")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(privBlock.Bytes)
	if err != nil {
		pk, err2 := x509.ParsePKCS1PrivateKey(privBlock.Bytes)
		if err2 != nil {
			return fmt.Errorf("failed to parse private key from env: %w", err)
		}
		privateKey = pk
	}

	rsaPriv, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("private key from env is not RSA")
	}

	pubBlock, _ := pem.Decode(pubPEM)
	if pubBlock == nil {
		return fmt.Errorf("failed to decode public key PEM block from env")
	}

	publicKey, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key from env: %w", err)
	}

	rsaPub, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("public key from env is not RSA")
	}

	km.mu.Lock()
	km.privateKey = rsaPriv
	km.publicKey = rsaPub
	km.mu.Unlock()

	return nil
}

// loadKeys reads PEM-encoded RSA keys from disk.
func (km *KeyManager) loadKeys(privateKeyPath, publicKeyPath string) error {
	// Load private key
	privPEM, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key file: %w", err)
	}

	privBlock, _ := pem.Decode(privPEM)
	if privBlock == nil {
		return fmt.Errorf("failed to decode private key PEM block")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(privBlock.Bytes)
	if err != nil {
		// Fallback to PKCS1
		pk, err2 := x509.ParsePKCS1PrivateKey(privBlock.Bytes)
		if err2 != nil {
			return fmt.Errorf("failed to parse private key (tried PKCS8 and PKCS1): %w", err)
		}
		privateKey = pk
	}

	rsaPriv, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("private key is not RSA")
	}

	// Load public key
	pubPEM, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key file: %w", err)
	}

	pubBlock, _ := pem.Decode(pubPEM)
	if pubBlock == nil {
		return fmt.Errorf("failed to decode public key PEM block")
	}

	publicKey, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("public key is not RSA")
	}

	km.mu.Lock()
	km.privateKey = rsaPriv
	km.publicKey = rsaPub
	km.mu.Unlock()

	return nil
}

// generateAndSaveKeys creates a new RSA-2048 key pair and persists to disk.
func (km *KeyManager) generateAndSaveKeys(privateKeyPath, publicKeyPath string) error {
	// Generate 2048-bit RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Ensure directories exist
	if err := os.MkdirAll(filepath.Dir(privateKeyPath), 0700); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(publicKeyPath), 0700); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	// Marshal private key to PKCS8 DER
	privDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	// Write private key PEM
	privFile, err := os.OpenFile(privateKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create private key file: %w", err)
	}
	defer privFile.Close()

	if err := pem.Encode(privFile, &pem.Block{Type: "PRIVATE KEY", Bytes: privDER}); err != nil {
		return fmt.Errorf("failed to write private key PEM: %w", err)
	}

	// Marshal public key to PKIX DER
	pubDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %w", err)
	}

	// Write public key PEM
	pubFile, err := os.OpenFile(publicKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create public key file: %w", err)
	}
	defer pubFile.Close()

	if err := pem.Encode(pubFile, &pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}); err != nil {
		return fmt.Errorf("failed to write public key PEM: %w", err)
	}

	km.mu.Lock()
	km.privateKey = privateKey
	km.publicKey = &privateKey.PublicKey
	km.mu.Unlock()

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
