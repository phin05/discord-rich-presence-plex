package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
)

const nonceLength = 12

// This is meant to be used only for plain eye obfuscation for credentials in the config, hence the hardcoded key
var key = []byte{24, 174, 1, 146, 150, 5, 51, 175, 232, 26, 179, 117, 240, 0, 161, 11, 26, 164, 191, 167, 92, 74, 158, 89, 99, 248, 234, 165, 118, 201, 28, 225}

func aesGcmEncrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new GCM: %w", err)
	}
	nonce := generateRandomBytes(nonceLength)
	encryptedData := gcm.Seal(nil, nonce, data, nil)
	payload := append(nonce, encryptedData...) //nolint:gocritic
	return payload, nil
}

func aesGcmDecrypt(payload []byte) ([]byte, error) {
	if len(payload) < nonceLength {
		return nil, errors.New("invalid payload")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new GCM: %w", err)
	}
	nonce := payload[:nonceLength]
	encryptedData := payload[nonceLength:]
	data, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("GCM open: %w", err)
	}
	return data, nil
}

func generateRandomBytes(length int) []byte {
	randomBytes := make([]byte, length)
	// rand.Read should never fail, but panic just in case
	if _, err := rand.Read(randomBytes); err != nil {
		panic(fmt.Errorf("read random bytes: %w", err))
	}
	return randomBytes
}
