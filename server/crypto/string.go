package crypto

import (
	"encoding/base64"
	"encoding/json/v2"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

const encryptionPrefix = "encrypted:"

type String struct {
	encrypted *string // Should be a nil pointer if decrypted is empty, for validate:"required" to work correctly
	decrypted string
}

func (s *String) Set(value string) {
	// Encrypted value will be set during marshaling, so only allocate an empty string for it here when setting a new decrypted value
	s.encrypted = new(string)
	s.decrypted = value
}

func (s *String) Value() string {
	return s.decrypted
}

func marshal[T any](s String, done func(encrypted string) (T, error)) (T, error) {
	if s.encrypted == nil || s.decrypted == "" {
		return done("")
	}
	encrypted := *s.encrypted
	if strings.HasPrefix(encrypted, encryptionPrefix) {
		return done(encrypted)
	}
	encryptedBytes, err := aesGcmEncrypt([]byte(s.decrypted))
	if err != nil {
		var zero T
		return zero, fmt.Errorf("encrypt: %w", err)
	}
	encrypted = encryptionPrefix + base64.StdEncoding.EncodeToString(encryptedBytes)
	*s.encrypted = encrypted
	return done(encrypted)
}

func (s String) MarshalJSON() ([]byte, error) {
	return marshal(s, func(encrypted string) ([]byte, error) {
		bytes, err := json.Marshal(encrypted)
		if err != nil {
			return nil, fmt.Errorf("marshal: %w", err)
		}
		return bytes, nil
	})
}

// Doesn't support pointer receiver
func (s String) MarshalYAML() (any, error) {
	return marshal(s, func(encrypted string) (any, error) {
		return encrypted, nil
	})
}

func (s *String) unmarshal(encrypted string) error {
	if encrypted == "" {
		return nil
	}
	if !strings.HasPrefix(encrypted, encryptionPrefix) {
		// Value is not encrypted so set it as the decrypted value directly
		s.Set(encrypted)
		return nil
	}
	s.encrypted = &encrypted
	encryptedBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(encrypted, encryptionPrefix))
	if err != nil {
		return fmt.Errorf("base64 decode: %w", err)
	}
	decryptedBytes, err := aesGcmDecrypt(encryptedBytes)
	if err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}
	s.decrypted = string(decryptedBytes)
	return nil
}

func (s *String) UnmarshalJSON(data []byte) error {
	var encrypted string
	if err := json.Unmarshal(data, &encrypted); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	return s.unmarshal(encrypted)
}

func (s *String) UnmarshalYAML(value *yaml.Node) error {
	return s.unmarshal(value.Value)
}
