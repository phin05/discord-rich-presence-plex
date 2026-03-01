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
	encrypted string
	decrypted string
}

func (s *String) Set(value string) error {
	if value == "" {
		s.encrypted = ""
	} else {
		encryptedBytes, err := aesGcmEncrypt([]byte(value))
		if err != nil {
			return err
		}
		s.encrypted = encryptionPrefix + base64.StdEncoding.EncodeToString(encryptedBytes)
	}
	s.decrypted = value
	return nil
}

func (s *String) Value() string {
	return s.decrypted
}

func (s *String) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.encrypted)
}

// Doesn't support pointer receiver
func (s String) MarshalYAML() (any, error) {
	return s.encrypted, nil
}

func (s *String) unmarshal(encrypted string) error {
	if encrypted == "" {
		return nil
	}
	if !strings.HasPrefix(encrypted, encryptionPrefix) {
		// Value is not encrypted so set it as the decrypted value directly
		if err := s.Set(encrypted); err != nil {
			return fmt.Errorf("set decrypted value: %w", err)
		}
		return nil
	}
	s.encrypted = encrypted
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
