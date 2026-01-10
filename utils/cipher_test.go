package utils

import (
	"strings"
	"testing"
)

func TestAesEncryptDecrypt(t *testing.T) {
	testCases := []struct {
		name      string
		data      string
		password  string
		shouldErr bool
	}{
		{
			name:      "Basic encryption and decryption",
			data:      "Hello, World!",
			password:  "password123",
			shouldErr: false,
		},
		{
			name:      "Long data",
			data:      strings.Repeat("This is a test message. ", 100),
			password:  "securepassword",
			shouldErr: false,
		},
		{
			name:      "Special characters in password",
			data:      "Sensitive data",
			password:  "p@$$w0rd!#$%^&*()",
			shouldErr: false,
		},
		{
			name:      "Empty data",
			data:      "",
			password:  "password",
			shouldErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test encryption
			encrypted, err := AesEncrypt([]byte(tc.data), tc.password)
			if tc.shouldErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("Unexpected error during encryption: %v", err)
			}
			if err != nil {
				return
			}

			// Verify encrypted data is different from original
			if encrypted == tc.data {
				t.Errorf("Encrypted data should be different from original")
			}

			// Verify encrypted data contains hex characters only
			if !isHexString(encrypted) {
				t.Errorf("Encrypted data should be hex encoded")
			}

			// Test decryption
			decrypted, err := AesDecrypt(encrypted, tc.password)
			if err != nil {
				t.Errorf("Decryption failed: %v", err)
			}

			// Verify decrypted data matches original
			if string(decrypted) != tc.data {
				t.Errorf("Decrypted data doesn't match original. Got: %s, Want: %s", string(decrypted), tc.data)
			}
		})
	}
}

func TestAesDecryptWrongPassword(t *testing.T) {
	data := "Secret message"
	password := "correctpassword"

	// Encrypt with correct password
	encrypted, err := AesEncrypt([]byte(data), password)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Try to decrypt with wrong password
	wrongPassword := "wrongpassword"
	_, err = AesDecrypt(encrypted, wrongPassword)
	if err == nil {
		t.Errorf("Expected error when decrypting with wrong password, got nil")
	}
}

func TestAesDecryptCorruptedData(t *testing.T) {
	corruptedData := "notvalidhex"
	password := "password"

	_, err := AesDecrypt(corruptedData, password)
	if err == nil {
		t.Errorf("Expected error for corrupted data, got nil")
	}

	// Valid hex but too short
	shortData := "abc123"
	_, err = AesDecrypt(shortData, password)
	if err == nil {
		t.Errorf("Expected error for too short data, got nil")
	}
}

func TestEncryptionDeterminism(t *testing.T) {
	data := "Test data"
	password := "password"

	// Encrypt the same data twice with the same password
	enc1, err1 := AesEncrypt([]byte(data), password)
	enc2, err2 := AesEncrypt([]byte(data), password)

	if err1 != nil || err2 != nil {
		t.Fatalf("Encryption failed: %v, %v", err1, err2)
	}

	// Encrypted data should be different due to random salt and nonce
	if enc1 == enc2 {
		t.Errorf("Encrypted data should be different due to random salt and nonce")
	}

	// But both should decrypt to the same original data
	dec1, _ := AesDecrypt(enc1, password)
	dec2, _ := AesDecrypt(enc2, password)

	if string(dec1) != data || string(dec2) != data {
		t.Errorf("Decrypted data doesn't match original")
	}
}

func TestPBKDF2KeyDerivation(t *testing.T) {
	password := "testpassword"
	salt1 := make([]byte, SaltSize)
	salt2 := make([]byte, SaltSize)

	// Generate different salts
	for i := range salt1 {
		salt1[i] = byte(i)
		salt2[i] = byte(i + 1)
	}

	key1 := deriveKey(password, salt1)
	key2 := deriveKey(password, salt2)

	// Same password with different salts should produce different keys
	if string(key1) == string(key2) {
		t.Errorf("Different salts should produce different keys")
	}

	// Same password with same salt should produce same key
	key3 := deriveKey(password, salt1)
	if string(key1) != string(key3) {
		t.Errorf("Same password and salt should produce same key")
	}
}

// Helper function to check if string is valid hex
func isHexString(s string) bool {
	if len(s)%2 != 0 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
