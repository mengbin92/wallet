package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"golang.org/x/crypto/pbkdf2"
	"github.com/pkg/errors"
)

const (
	// SaltSize is the size of the salt in bytes
	SaltSize = 32
	// PBKDF2Iterations is the number of iterations for PBKDF2
	PBKDF2Iterations = 100000
	// KeySize is the size of the derived key in bytes
	KeySize = 32
)

// AesEncrypt encrypts the given data using the provided passphrase with PBKDF2 key derivation
func AesEncrypt(data []byte, passphrase string) (string, error) {
	// Generate a random salt
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", errors.Wrap(err, "failed to generate salt")
	}

	// Derive key using PBKDF2
	key := deriveKey(passphrase, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.Wrap(err, "failed to create cipher block")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.Wrap(err, "failed to create gcm block")
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", errors.Wrap(err, "failed to generate nonce")
	}

	// Seal: salt + nonce + ciphertext
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	// Prepend salt to ciphertext
	result := append(salt, ciphertext...)

	return hex.EncodeToString(result), nil
}

// AesDecrypt decrypts the given data using the provided passphrase with PBKDF2 key derivation
func AesDecrypt(encryptedData string, passphrase string) ([]byte, error) {
	data, err := hex.DecodeString(encryptedData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode encrypted data")
	}

	// Check minimum length: salt (32) + nonce (12) + some ciphertext
	if len(data) < SaltSize+12 {
		return nil, errors.New("ciphertext too short")
	}

	// Extract salt
	salt := data[:SaltSize]
	ciphertext := data[SaltSize:]

	// Derive key using PBKDF2 with the extracted salt
	key := deriveKey(passphrase, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cipher block")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gcm block")
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short after removing salt")
	}

	nonce, encrypted := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt ciphertext (wrong passphrase or corrupted data)")
	}

	return plaintext, nil
}

// deriveKey derives a key from passphrase using PBKDF2
func deriveKey(passphrase string, salt []byte) []byte {
	return pbkdf2.Key([]byte(passphrase), salt, PBKDF2Iterations, KeySize, sha256.New)
}
