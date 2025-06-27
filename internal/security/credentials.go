package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/term"
)

// CredentialStore represents a secure credential storage system
type CredentialStore struct {
	StorePath string
	MasterKey []byte
	Salt      []byte
	encrypted map[string]EncryptedCredential
}

// EncryptedCredential represents encrypted credential data
type EncryptedCredential struct {
	Data      string    `json:"data"`
	Nonce     string    `json:"nonce"`
	CreatedAt time.Time `json:"created_at"`
	LastUsed  time.Time `json:"last_used"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// Credential represents decrypted credential data
type Credential struct {
	Username    string            `json:"username"`
	Password    string            `json:"password"`
	AuthType    string            `json:"auth_type"`
	Server      string            `json:"server"`
	Port        int               `json:"port"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// NewCredentialStore creates a new credential store
func NewCredentialStore(storePath string) (*CredentialStore, error) {
	// Ensure the directory exists
	dir := filepath.Dir(storePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create store directory: %v", err)
	}

	store := &CredentialStore{
		StorePath: storePath,
		encrypted: make(map[string]EncryptedCredential),
	}

	// Try to load existing store
	if _, err := os.Stat(storePath); err == nil {
		if err := store.loadStore(); err != nil {
			return nil, fmt.Errorf("failed to load existing store: %v", err)
		}
	}

	return store, nil
}

// InitializeMasterKey initializes or prompts for the master key
func (cs *CredentialStore) InitializeMasterKey() error {
	if cs.MasterKey != nil {
		return nil // Already initialized
	}

	// Check if we have a stored salt
	if cs.Salt == nil {
		// Generate new salt
		cs.Salt = make([]byte, 32)
		if _, err := rand.Read(cs.Salt); err != nil {
			return fmt.Errorf("failed to generate salt: %v", err)
		}
	}

	// Prompt for master password
	fmt.Print("Enter master password for credential store: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("failed to read password: %v", err)
	}

	// Derive key from password
	cs.MasterKey = pbkdf2.Key(password, cs.Salt, 100000, 32, sha256.New)

	// Clear password from memory
	for i := range password {
		password[i] = 0
	}

	return nil
}

// SetMasterKey sets the master key directly (for testing or automated use)
func (cs *CredentialStore) SetMasterKey(password string) error {
	if cs.Salt == nil {
		cs.Salt = make([]byte, 32)
		if _, err := rand.Read(cs.Salt); err != nil {
			return fmt.Errorf("failed to generate salt: %v", err)
		}
	}

	cs.MasterKey = pbkdf2.Key([]byte(password), cs.Salt, 100000, 32, sha256.New)
	return nil
}

// StoreCredential stores a credential securely
func (cs *CredentialStore) StoreCredential(name string, cred *Credential) error {
	if cs.MasterKey == nil {
		return errors.New("master key not initialized")
	}

	// Serialize credential
	data, err := json.Marshal(cred)
	if err != nil {
		return fmt.Errorf("failed to serialize credential: %v", err)
	}

	// Encrypt data
	encryptedData, nonce, err := cs.encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt credential: %v", err)
	}

	// Store encrypted credential
	cs.encrypted[name] = EncryptedCredential{
		Data:      base64.StdEncoding.EncodeToString(encryptedData),
		Nonce:     base64.StdEncoding.EncodeToString(nonce),
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
	}

	// Save to disk
	return cs.saveStore()
}

// GetCredential retrieves and decrypts a credential
func (cs *CredentialStore) GetCredential(name string) (*Credential, error) {
	if cs.MasterKey == nil {
		return nil, errors.New("master key not initialized")
	}

	encCred, exists := cs.encrypted[name]
	if !exists {
		return nil, fmt.Errorf("credential '%s' not found", name)
	}

	// Check expiration
	if !encCred.ExpiresAt.IsZero() && time.Now().After(encCred.ExpiresAt) {
		return nil, fmt.Errorf("credential '%s' has expired", name)
	}

	// Decode base64
	data, err := base64.StdEncoding.DecodeString(encCred.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted data: %v", err)
	}

	nonce, err := base64.StdEncoding.DecodeString(encCred.Nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decode nonce: %v", err)
	}

	// Decrypt data
	decryptedData, err := cs.decrypt(data, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential: %v", err)
	}

	// Deserialize credential
	var cred Credential
	if err := json.Unmarshal(decryptedData, &cred); err != nil {
		return nil, fmt.Errorf("failed to deserialize credential: %v", err)
	}

	// Update last used time
	encCred.LastUsed = time.Now()
	cs.encrypted[name] = encCred
	cs.saveStore()

	return &cred, nil
}

// ListCredentials returns a list of stored credential names
func (cs *CredentialStore) ListCredentials() []string {
	names := make([]string, 0, len(cs.encrypted))
	for name := range cs.encrypted {
		names = append(names, name)
	}
	return names
}

// DeleteCredential removes a credential from the store
func (cs *CredentialStore) DeleteCredential(name string) error {
	if _, exists := cs.encrypted[name]; !exists {
		return fmt.Errorf("credential '%s' not found", name)
	}

	delete(cs.encrypted, name)
	return cs.saveStore()
}

// SetExpiration sets an expiration time for a credential
func (cs *CredentialStore) SetExpiration(name string, expiresAt time.Time) error {
	encCred, exists := cs.encrypted[name]
	if !exists {
		return fmt.Errorf("credential '%s' not found", name)
	}

	encCred.ExpiresAt = expiresAt
	cs.encrypted[name] = encCred
	return cs.saveStore()
}

// encrypt encrypts data using AES-GCM
func (cs *CredentialStore) encrypt(data []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(cs.MasterKey)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, data, nil)
	return ciphertext, nonce, nil
}

// decrypt decrypts data using AES-GCM
func (cs *CredentialStore) decrypt(ciphertext, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(cs.MasterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, nonce, ciphertext, nil)
}

// saveStore saves the encrypted store to disk
func (cs *CredentialStore) saveStore() error {
	storeData := struct {
		Salt      string                         `json:"salt"`
		Encrypted map[string]EncryptedCredential `json:"encrypted"`
	}{
		Salt:      base64.StdEncoding.EncodeToString(cs.Salt),
		Encrypted: cs.encrypted,
	}

	data, err := json.MarshalIndent(storeData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize store: %v", err)
	}

	// Write with restricted permissions
	return os.WriteFile(cs.StorePath, data, 0600)
}

// loadStore loads the encrypted store from disk
func (cs *CredentialStore) loadStore() error {
	data, err := os.ReadFile(cs.StorePath)
	if err != nil {
		return fmt.Errorf("failed to read store file: %v", err)
	}

	var storeData struct {
		Salt      string                         `json:"salt"`
		Encrypted map[string]EncryptedCredential `json:"encrypted"`
	}

	if err := json.Unmarshal(data, &storeData); err != nil {
		return fmt.Errorf("failed to deserialize store: %v", err)
	}

	cs.Salt, err = base64.StdEncoding.DecodeString(storeData.Salt)
	if err != nil {
		return fmt.Errorf("failed to decode salt: %v", err)
	}

	cs.encrypted = storeData.Encrypted
	return nil
}

// RotateMasterKey changes the master key and re-encrypts all credentials
func (cs *CredentialStore) RotateMasterKey(newPassword string) error {
	if cs.MasterKey == nil {
		return errors.New("master key not initialized")
	}

	// Store all credentials in memory (decrypted)
	credentials := make(map[string]*Credential)
	for name := range cs.encrypted {
		cred, err := cs.GetCredential(name)
		if err != nil {
			return fmt.Errorf("failed to decrypt credential %s during key rotation: %v", name, err)
		}
		credentials[name] = cred
	}

	// Generate new salt and derive new key
	newSalt := make([]byte, 32)
	if _, err := rand.Read(newSalt); err != nil {
		return fmt.Errorf("failed to generate new salt: %v", err)
	}

	newKey := pbkdf2.Key([]byte(newPassword), newSalt, 100000, 32, sha256.New)

	// Update store with new key and salt
	cs.MasterKey = newKey
	cs.Salt = newSalt
	cs.encrypted = make(map[string]EncryptedCredential)

	// Re-encrypt all credentials with new key
	for name, cred := range credentials {
		if err := cs.StoreCredential(name, cred); err != nil {
			return fmt.Errorf("failed to re-encrypt credential %s: %v", name, err)
		}
	}

	return nil
}

// ValidateIntegrity checks if the store can be properly decrypted
func (cs *CredentialStore) ValidateIntegrity() error {
	if cs.MasterKey == nil {
		return errors.New("master key not initialized")
	}

	for name := range cs.encrypted {
		_, err := cs.GetCredential(name)
		if err != nil {
			return fmt.Errorf("integrity check failed for credential %s: %v", name, err)
		}
	}

	return nil
}

// Cleanup removes expired credentials
func (cs *CredentialStore) Cleanup() error {
	now := time.Now()
	expiredCredentials := make([]string, 0)

	for name, encCred := range cs.encrypted {
		if !encCred.ExpiresAt.IsZero() && now.After(encCred.ExpiresAt) {
			expiredCredentials = append(expiredCredentials, name)
		}
	}

	for _, name := range expiredCredentials {
		delete(cs.encrypted, name)
	}

	if len(expiredCredentials) > 0 {
		return cs.saveStore()
	}

	return nil
}
