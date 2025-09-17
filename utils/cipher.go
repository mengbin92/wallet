package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/pkg/errors"
)

// AesEncrypt encrypts the given data using the provided passphrase
func AesEncrypt(data []byte, passphrase string) (string, error) {
	block, err := aes.NewCipher(createHash(passphrase))
	if err != nil {
		return "", errors.Wrapf(err, "failed to create cipher block (passphrase: %s)", passphrase)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create gcm block (passphrase: %s)", passphrase)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", errors.Wrapf(err, "failed to generate nonce (passphrase: %s)", passphrase)
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return hex.EncodeToString(ciphertext), nil
}

// AesDecrypt decrypts the given data using the provided passphrase
func AesDecrypt(encryptedData string, passphrase string) ([]byte, error) {
	data, err := hex.DecodeString(encryptedData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode encrypted data")
	}

	block, err := aes.NewCipher(createHash(passphrase))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cipher block")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gcm block")
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt ciphertext")
	}

	return plaintext, nil
}

// createHash creates a hash of the passphrase
func createHash(passphrase string) []byte {
	hash := sha256.New()
	hash.Write([]byte(passphrase))
	return hash.Sum(nil)
}
