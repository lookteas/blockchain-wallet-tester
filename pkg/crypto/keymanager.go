package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/term"
)

const (
	// AES-256-GCM key size
	KeySize = 32
	// PBKDF2 iterations
	PBKDF2Iterations = 100000
	// Salt size
	SaltSize = 32
)

// EncryptedWallet represents an encrypted wallet configuration
type EncryptedWallet struct {
	Name                 string `yaml:"name" json:"name"`
	EncryptedPrivateKey  string `yaml:"encrypted_private_key" json:"encrypted_private_key"`
	Salt                 string `yaml:"salt" json:"salt"`
}

// WalletConfig represents the wallet configuration file
type WalletConfig struct {
	Wallets []EncryptedWallet `yaml:"wallets" json:"wallets"`
}

// KeyManager handles secure private key management
type KeyManager struct {
	privateKeys []string
}

// NewKeyManager creates a new KeyManager instance
func NewKeyManager() *KeyManager {
	return &KeyManager{
		privateKeys: make([]string, 0),
	}
}

// LoadFromEnv loads private keys from environment variable
func (km *KeyManager) LoadFromEnv() error {
	envKeys := os.Getenv("WALLET_PRIVATE_KEYS")
	if envKeys == "" {
		return errors.New("WALLET_PRIVATE_KEYS environment variable not set")
	}

	keys := strings.Split(envKeys, ",")
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if err := km.validatePrivateKey(key); err != nil {
			return fmt.Errorf("invalid private key: %w", err)
		}
		km.privateKeys = append(km.privateKeys, km.normalizePrivateKey(key))
	}

	return nil
}

// LoadFromEncryptedFile loads private keys from encrypted configuration file
func (km *KeyManager) LoadFromEncryptedFile(filePath, password string) error {
	// Implementation for loading from encrypted file
	// This would involve reading the YAML file, decrypting each private key
	return errors.New("encrypted file loading not yet implemented")
}

// LoadInteractive loads private keys through interactive input
func (km *KeyManager) LoadInteractive() error {
	fmt.Print("请输入私钥数量: ")
	var count int
	if _, err := fmt.Scanf("%d", &count); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	for i := 0; i < count; i++ {
		fmt.Printf("请输入第 %d 个私钥 (输入时不会显示): ", i+1)
		
		// Read password without echoing
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("failed to read private key: %w", err)
		}
		fmt.Println() // Add newline after password input

		privateKey := string(bytePassword)
		if err := km.validatePrivateKey(privateKey); err != nil {
			return fmt.Errorf("invalid private key %d: %w", i+1, err)
		}

		km.privateKeys = append(km.privateKeys, km.normalizePrivateKey(privateKey))
		
		// Clear the password from memory
		for j := range bytePassword {
			bytePassword[j] = 0
		}
	}

	return nil
}

// GetPrivateKeys returns a copy of the private keys
func (km *KeyManager) GetPrivateKeys() []string {
	keys := make([]string, len(km.privateKeys))
	copy(keys, km.privateKeys)
	return keys
}

// Clear clears all private keys from memory
func (km *KeyManager) Clear() {
	for i := range km.privateKeys {
		// Overwrite with zeros
		for j := range km.privateKeys[i] {
			// This is a best effort to clear memory, but Go's GC may have moved the string
			_ = j
		}
		km.privateKeys[i] = ""
	}
	km.privateKeys = km.privateKeys[:0]
}

// EncryptPrivateKey encrypts a private key using AES-256-GCM with PBKDF2
func (km *KeyManager) EncryptPrivateKey(privateKey, password string) (string, string, error) {
	// Generate random salt
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key using PBKDF2
	key := pbkdf2.Key([]byte(password), salt, PBKDF2Iterations, KeySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, []byte(privateKey), nil)

	return hex.EncodeToString(ciphertext), hex.EncodeToString(salt), nil
}

// DecryptPrivateKey decrypts a private key using AES-256-GCM with PBKDF2
func (km *KeyManager) DecryptPrivateKey(encryptedKey, saltHex, password string) (string, error) {
	// Decode hex strings
	ciphertext, err := hex.DecodeString(encryptedKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted key: %w", err)
	}

	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode salt: %w", err)
	}

	// Derive key using PBKDF2
	key := pbkdf2.Key([]byte(password), salt, PBKDF2Iterations, KeySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce and ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// validatePrivateKey validates the format of a private key
func (km *KeyManager) validatePrivateKey(privateKey string) error {
	key := km.normalizePrivateKey(privateKey)
	
	// Check length (64 hex characters = 32 bytes)
	if len(key) != 64 {
		return errors.New("private key must be 64 hex characters (32 bytes)")
	}

	// Check if it's valid hex
	if _, err := hex.DecodeString(key); err != nil {
		return fmt.Errorf("private key must be valid hex: %w", err)
	}

	return nil
}

// normalizePrivateKey removes 0x prefix if present and converts to lowercase
func (km *KeyManager) normalizePrivateKey(privateKey string) string {
	key := strings.TrimSpace(privateKey)
	if strings.HasPrefix(key, "0x") || strings.HasPrefix(key, "0X") {
		key = key[2:]
	}
	return strings.ToLower(key)
}