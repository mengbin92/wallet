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
