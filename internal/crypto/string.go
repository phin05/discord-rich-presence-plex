package crypto

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

const encryptionPrefix = "encrypted:"

type String struct {
	encrypted *string
	decrypted string
}

func (s *String) Set(value string) {
	s.encrypted = new(string)
	s.decrypted = value
}

func (s *String) Value() string {
	return s.decrypted
}

func marshal[T any](s String, done func(string) (T, error)) (T, error) {
	if s.encrypted == nil || s.decrypted == "" {
		return done("")
	}
	if !strings.HasPrefix(*s.encrypted, encryptionPrefix) {
		encryptedBytes, err := aesGcmEncrypt([]byte(s.decrypted))
		if err != nil {
			var zero T
			return zero, fmt.Errorf("encrypt: %w", err)
		}
		*s.encrypted = encryptionPrefix + base64.StdEncoding.EncodeToString(encryptedBytes)
	}
	return done(*s.encrypted)
}

func (s String) MarshalJSON() ([]byte, error) {
	return marshal(s, func(encrypted string) ([]byte, error) {
		return json.Marshal(encrypted)
	})
}

func (s String) MarshalYAML() (any, error) {
	return marshal(s, func(encrypted string) (any, error) {
		return encrypted, nil
	})
}

func unmarshal(s *String, encrypted string) error {
	s.encrypted = new(string)
	*s.encrypted = encrypted
	if !strings.HasPrefix(encrypted, encryptionPrefix) {
		s.decrypted = encrypted
		return nil
	}
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
	return unmarshal(s, encrypted)
}

func (s *String) UnmarshalYAML(value *yaml.Node) error {
	return unmarshal(s, value.Value)
}
