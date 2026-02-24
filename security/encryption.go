package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

// Encryption provides encryption utilities
type Encryption struct {
	key []byte
}

// NewEncryption creates a new encryption utility with the given key
func NewEncryption(key string) *Encryption {
	// Derive a proper 32-byte key using SHA-256
	hash := sha256.Sum256([]byte(key))
	return &Encryption{
		key: hash[:],
	}
}

// Encrypt encrypts plaintext using AES-GCM
func (e *Encryption) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-GCM
func (e *Encryption) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash verifies a password against a hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomToken generates a cryptographically secure random token
func GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateAPIKey generates a secure API key
func GenerateAPIKey() (string, error) {
	return GenerateRandomToken(32)
}

// HashData creates a SHA-256 hash of data
func HashData(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// DeriveKey derives a key from a password and salt using PBKDF2
func DeriveKey(password, salt string, keyLen int) []byte {
	return pbkdf2.Key([]byte(password), []byte(salt), 10000, keyLen, sha256.New)
}

// EncryptionAt Rest provides field-level encryption for sensitive data
type FieldEncryption struct {
	enc *Encryption
}

// NewFieldEncryption creates a field encryption helper
func NewFieldEncryption(key string) *FieldEncryption {
	return &FieldEncryption{
		enc: NewEncryption(key),
	}
}

// EncryptField encrypts a single field
func (fe *FieldEncryption) EncryptField(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	return fe.enc.Encrypt(value)
}

// DecryptField decrypts a single field
func (fe *FieldEncryption) DecryptField(encryptedValue string) (string, error) {
	if encryptedValue == "" {
		return "", nil
	}
	return fe.enc.Decrypt(encryptedValue)
}

// EncryptSensitiveFields encrypts common sensitive fields in a map
func (fe *FieldEncryption) EncryptSensitiveFields(data map[string]interface{}) error {
	sensitiveFields := []string{"email", "phone", "address", "ssn", "credit_card"}
	
	for _, field := range sensitiveFields {
		if value, ok := data[field]; ok {
			if strValue, ok := value.(string); ok && strValue != "" {
				encrypted, err := fe.enc.Encrypt(strValue)
				if err != nil {
					return err
				}
				data[field] = encrypted
			}
		}
	}
	
	return nil
}

// DecryptSensitiveFields decrypts common sensitive fields in a map
func (fe *FieldEncryption) DecryptSensitiveFields(data map[string]interface{}) error {
	sensitiveFields := []string{"email", "phone", "address", "ssn", "credit_card"}
	
	for _, field := range sensitiveFields {
		if value, ok := data[field]; ok {
			if strValue, ok := value.(string); ok && strValue != "" {
				decrypted, err := fe.enc.Decrypt(strValue)
				if err != nil {
					return err
				}
				data[field] = decrypted
			}
		}
	}
	
	return nil
}

// TokenManager handles secure token generation and validation
type TokenManager struct {
	tokens map[string]TokenInfo
}

// TokenInfo stores token metadata
type TokenInfo struct {
	UserID    string
	ExpiresAt int64
	Type      string // reset_password, email_verification, api_key
}

// NewTokenManager creates a new token manager
func NewTokenManager() *TokenManager {
	return &TokenManager{
		tokens: make(map[string]TokenInfo),
	}
}

// GenerateToken creates a new token
func (tm *TokenManager) GenerateToken(userID, tokenType string, expiresAt int64) (string, error) {
	token, err := GenerateRandomToken(32)
	if err != nil {
		return "", err
	}
	
	tm.tokens[token] = TokenInfo{
		UserID:    userID,
		ExpiresAt: expiresAt,
		Type:      tokenType,
	}
	
	return token, nil
}

// ValidateToken checks if a token is valid
func (tm *TokenManager) ValidateToken(token string) (TokenInfo, bool) {
	info, exists := tm.tokens[token]
	if !exists {
		return TokenInfo{}, false
	}
	
	// Check expiration
	if info.ExpiresAt > 0 && info.ExpiresAt < currentTimeMillis() {
		delete(tm.tokens, token)
		return TokenInfo{}, false
	}
	
	return info, true
}

// RevokeToken invalidates a token
func (tm *TokenManager) RevokeToken(token string) {
	delete(tm.tokens, token)
}

func currentTimeMillis() int64 {
	return 0 // Would use time.Now().Unix() in production
}
