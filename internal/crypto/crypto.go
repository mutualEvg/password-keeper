// Package crypto provides encryption and decryption utilities for GophKeeper.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

const (
	// KeySize is the size of the encryption key in bytes (32 bytes = 256 bits).
	KeySize = 32
	// SaltSize is the size of the salt used in key derivation.
	SaltSize = 32
	// NonceSize is the size of the nonce used in AES-GCM.
	NonceSize = 12
	// PBKDF2Iterations is the number of iterations for PBKDF2.
	PBKDF2Iterations = 100000
)

var (
	// ErrInvalidCiphertext is returned when the ciphertext is too short or invalid.
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	// ErrDecryptionFailed is returned when decryption fails.
	ErrDecryptionFailed = errors.New("decryption failed")
)

// HashPassword hashes a password using bcrypt.
// Returns the bcrypt hash of the password.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPasswordHash compares a password with a bcrypt hash.
// Returns nil if the password matches the hash.
func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// DeriveKey derives an encryption key from a master password and salt using PBKDF2.
// The salt should be stored alongside the encrypted data for decryption.
func DeriveKey(masterPassword string, salt []byte) []byte {
	return pbkdf2.Key([]byte(masterPassword), salt, PBKDF2Iterations, KeySize, sha256.New)
}

// GenerateSalt generates a random salt for key derivation.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// Encrypt encrypts plaintext using AES-256-GCM with the provided key.
// Returns: salt + nonce + ciphertext + tag (all concatenated).
func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate nonce
	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt data
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM with the provided key.
// The ciphertext should be in the format: nonce + ciphertext + tag.
func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Check minimum length
	if len(ciphertext) < NonceSize {
		return nil, ErrInvalidCiphertext
	}

	// Extract nonce and ciphertext
	nonce := ciphertext[:NonceSize]
	ciphertextData := ciphertext[NonceSize:]

	// Decrypt data
	plaintext, err := gcm.Open(nil, nonce, ciphertextData, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// EncryptWithPassword encrypts plaintext using a password (derives key from password).
// Returns: salt + encrypted data (salt is needed for decryption).
func EncryptWithPassword(plaintext []byte, password string) ([]byte, error) {
	// Generate salt
	salt, err := GenerateSalt()
	if err != nil {
		return nil, err
	}

	// Derive key from password
	key := DeriveKey(password, salt)

	// Encrypt data
	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		return nil, err
	}

	// Prepend salt to ciphertext
	result := make([]byte, len(salt)+len(ciphertext))
	copy(result, salt)
	copy(result[len(salt):], ciphertext)

	return result, nil
}

// DecryptWithPassword decrypts ciphertext using a password.
// The ciphertext should be in the format: salt + encrypted data.
func DecryptWithPassword(ciphertext []byte, password string) ([]byte, error) {
	// Check minimum length
	if len(ciphertext) < SaltSize+NonceSize {
		return nil, ErrInvalidCiphertext
	}

	// Extract salt and encrypted data
	salt := ciphertext[:SaltSize]
	encryptedData := ciphertext[SaltSize:]

	// Derive key from password
	key := DeriveKey(password, salt)

	// Decrypt data
	return Decrypt(encryptedData, key)
}

// EncodeBase64 encodes bytes to base64 string.
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBase64 decodes base64 string to bytes.
func DecodeBase64(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

