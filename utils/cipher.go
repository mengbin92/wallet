package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

// AesEncrypt encrypts the given data using the provided passphrase
func AesEncrypt(data []byte, passphrase string) (string, error) {
	block, err := aes.NewCipher(createHash(passphrase))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return hex.EncodeToString(ciphertext), nil
}

// AesDecrypt decrypts the given data using the provided passphrase
func AesDecrypt(encryptedData string, passphrase string) ([]byte, error) {
	data, err := hex.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(createHash(passphrase))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// createHash creates a hash of the passphrase
func createHash(passphrase string) []byte {
	hash := sha256.New()
	hash.Write([]byte(passphrase))
	return hash.Sum(nil)
}
