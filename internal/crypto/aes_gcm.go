package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

const nonceLength = 12

// This is meant to be used only for plain eye obfuscation for credentials in the config, hence the hardcoded key
var key = []byte{24, 174, 1, 146, 150, 5, 51, 175, 232, 26, 179, 117, 240, 0, 161, 11, 26, 164, 191, 167, 92, 74, 158, 89, 99, 248, 234, 165, 118, 201, 28, 225}

func aesGcmEncrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce, err := generateRandomBytes(nonceLength)
	if err != nil {
		return nil, err
	}
	encryptedData := gcm.Seal(nil, nonce, data, nil)
	payload := append(nonce, encryptedData...)
	return payload, nil
}

func aesGcmDecrypt(data []byte) ([]byte, error) {
	if len(data) < nonceLength {
		return nil, fmt.Errorf("invalid payload")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := data[:nonceLength]
	encryptedData := data[nonceLength:]
	return gcm.Open(nil, nonce, encryptedData, nil)
}

func generateRandomBytes(length int) ([]byte, error) {
	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("read random bytes: %w", err)
	}
	return randomBytes, nil
}
