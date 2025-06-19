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
			t.Fatalf("encrypt: %s", err)
		}
		decryptedData, err := aesGcmDecrypt(encryptedData)
		if err != nil {
			t.Fatalf("decrypt: %s", err)
		}
		if !bytes.Equal(data, decryptedData) {
			t.Fatal("decrypted data differs from original data")
		}
	})
	t.Run("Invalid Payload", func(t *testing.T) {
		encryptedData := []byte{4, 8, 15, 16, 23, 42}
		_, err := aesGcmDecrypt(encryptedData)
		if err == nil {
			t.Fatal("decrypt succeeded with invalid payload")
		}
	})
	t.Run("Modified Payload", func(t *testing.T) {
		encryptedData, err := aesGcmEncrypt(data)
		if err != nil {
			t.Fatalf("encrypt: %s", err)
		}
		encryptedData[len(encryptedData)/2]++
		_, err = aesGcmDecrypt(encryptedData)
		if err == nil {
			t.Fatal("decrypt succeeded with modified payload")
		}
	})
	t.Run("Non-Deterministic Encryption", func(t *testing.T) {
		encryptedData1, err := aesGcmEncrypt(data)
		if err != nil {
			t.Fatalf("encrypt: %s", err)
		}
		encryptedData2, err := aesGcmEncrypt(data)
		if err != nil {
			t.Fatalf("encrypt: %s", err)
		}
		if bytes.Equal(encryptedData1, encryptedData2) {
			t.Fatal("encrypted outputs are the same")
		}
	})
}
