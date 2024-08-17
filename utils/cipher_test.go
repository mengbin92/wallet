package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptAndDecrypt(t *testing.T) {

	tests := []struct {
		name        string
		plainText   string
		encryptWord string
		decryptWord string
		success     bool
	}{
		{
			name:        "TestEncryptAndDecrypt success",
			plainText:   "hello world",
			encryptWord: "pwdtest0pwdtest0",
			decryptWord: "pwdtest0pwdtest0",
			success:     true,
		},
		{
			name:        "TestEncryptAndDecrypt fail",
			plainText:   "hello world",
			encryptWord: "pwdtest0pwdtest0",
			decryptWord: "pwdtest0pwdtest1",
			success:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.success {
				cipherText, err := AesEncrypt([]byte(tt.plainText), tt.encryptWord)
				assert.Empty(t, err)
				decryptedText, err := AesDecrypt(cipherText, tt.decryptWord)
				assert.Empty(t, err)
				assert.Equal(t, tt.plainText, string(decryptedText))
			} else {
				cipherText, err := AesEncrypt([]byte(tt.plainText), tt.encryptWord)
				assert.Empty(t, err)
				_, err = AesDecrypt(cipherText, tt.decryptWord)
				assert.NotEmpty(t, err)
				assert.Contains(t, err.Error(), "cipher: message authentication failed")
			}

		})
	}
}

func TestBIP38(t *testing.T) {
	tests := []struct {
		name        string
		plainText   string
		encryptWord string
		decryptWord string
		network     string
		success     bool
	}{
		{
			name:        "TestNet3Params TestBIP38Encrypt success",
			plainText:   "cUk3XmNvBjrRijzrZjJLoG9pUhhTGzZsQE4mBCUjdGXNnQDWHYXA",
			encryptWord: "pwdtest0pwdtest0",
			decryptWord: "pwdtest0pwdtest0",
			network:     "testnet",
			success:     true,
		},
		{
			name:        "TestNet3Params TestBIP38Encrypt fail",
			plainText:   "cUk3XmNvBjrRijzrZjJLoG9pUhhTGzZsQE4mBCUjdGXNnQDWHYXA",
			encryptWord: "pwdtest0pwdtest0",
			decryptWord: "pwdtest0pwdtest1",
			network:     "testnet",
			success:     false,
		},
		{
			name:        "MainNetParams TestBIP38Encrypt success",
			plainText:   "KxmE3EPbWGQNWZFR5iLgnr83Uev6AgaR2SEuxKVc6d5aDPgyCNKL",
			encryptWord: "pwdtest0pwdtest0",
			decryptWord: "pwdtest0pwdtest0",
			network:     "mainnet",
			success:     true,
		},
		{
			name:        "MainNetParams TestBIP38Encrypt fail",
			plainText:   "KxmE3EPbWGQNWZFR5iLgnr83Uev6AgaR2SEuxKVc6d5aDPgyCNKL",
			encryptWord: "pwdtest0pwdtest0",
			decryptWord: "pwdtest0pwdtest1",
			network:     "mainnet",
			success:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.success {
				cipherText, err := BIP38Encrypt(tt.plainText, tt.encryptWord)
				assert.Empty(t, err)
				decryptedText, err := BIP38Decrypt(cipherText, tt.decryptWord, tt.network)
				assert.Empty(t, err)
				assert.Equal(t, tt.plainText, decryptedText)
			} else {
				cipherText, err := BIP38Encrypt(tt.plainText, tt.encryptWord)
				assert.Empty(t, err)
				decryptText, err := BIP38Decrypt(cipherText, tt.decryptWord, tt.network)
				assert.Empty(t, err)
				assert.NotEqual(t, decryptText, tt.plainText)
			}

		})
	}
}
