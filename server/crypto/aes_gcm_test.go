package crypto

import (
	"bytes"
	"testing"
)

func TestAesGcm(t *testing.T) {
	data := []byte("Test data")
	t.Run("Encrypt and Decrypt", func(t *testing.T) {
		encryptedData, err := aesGcmEncrypt(data)
		if err != nil {
			t.Fatalf("encrypt: %v", err)
		}
		decryptedData, err := aesGcmDecrypt(encryptedData)
		if err != nil {
			t.Fatalf("decrypt: %v", err)
		}
		if !bytes.Equal(data, decryptedData) {
			t.Error("expected decrypted data to match original data")
		}
	})
	t.Run("Invalid Payload", func(t *testing.T) {
		encryptedData := []byte{4, 8, 15, 16, 23, 42}
		if _, err := aesGcmDecrypt(encryptedData); err == nil {
			t.Error("expected decryption of invalid payload to fail")
		}
	})
	t.Run("Modified Payload", func(t *testing.T) {
		encryptedData, err := aesGcmEncrypt(data)
		if err != nil {
			t.Fatalf("encrypt: %v", err)
		}
		encryptedData[len(encryptedData)/2]++
		if _, err := aesGcmDecrypt(encryptedData); err == nil {
			t.Error("expected decryption of modified payload to fail")
		}
	})
	t.Run("Non-Deterministic Encryption", func(t *testing.T) {
		encryptedData1, err := aesGcmEncrypt(data)
		if err != nil {
			t.Fatalf("encrypt: %v", err)
		}
		encryptedData2, err := aesGcmEncrypt(data)
		if err != nil {
			t.Fatalf("encrypt: %v", err)
		}
		if bytes.Equal(encryptedData1, encryptedData2) {
			t.Error("expected different outputs for multiple encryptions of the same data")
		}
	})
}
