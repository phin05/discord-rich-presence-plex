package config

import (
	"drpp/internal/crypto"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

const encryptionPrefix = "encrypted:"

type slice[T any] []T

func (s slice[T]) MarshalJSON() ([]byte, error) {
	if s == nil {
		return json.Marshal([]T{})
	}
	return json.Marshal([]T(s))
}

type Config struct {
	Web     Web         `yaml:"web" json:"web"`
	Display Display     `yaml:"display" json:"display"`
	Users   slice[User] `yaml:"users" json:"users"`
}

type Web struct {
	BindAddress string `yaml:"bindAddress" json:"bindAddress"`
	BindPort    int    `yaml:"bindPort" json:"bindPort"`
}

type Display struct {
	Duration     bool          `yaml:"duration" json:"duration"`
	Genres       bool          `yaml:"genres" json:"genres"`
	Album        bool          `yaml:"album" json:"album"`
	AlbumImage   bool          `yaml:"albumImage" json:"albumImage"`
	Artist       bool          `yaml:"artist" json:"artist"`
	ArtistImage  bool          `yaml:"artistImage" json:"artistImage"`
	Year         bool          `yaml:"year" json:"year"`
	StatusIcon   bool          `yaml:"statusIcon" json:"statusIcon"`
	ProgressMode string        `yaml:"progressMode" json:"progressMode"`
	Paused       bool          `yaml:"paused" json:"paused"`
	Posters      Posters       `yaml:"posters" json:"posters"`
	Buttons      slice[Button] `yaml:"buttons" json:"buttons"`
}

type Posters struct {
	Enabled       bool   `yaml:"enabled" json:"enabled"`
	ImgurClientID string `yaml:"imgurClientID" json:"imgurClientID"`
	MaxSize       int    `yaml:"maxSize" json:"maxSize"`
	Fit           bool   `yaml:"fit" json:"fit"`
}

type Button struct {
	Label      string        `yaml:"label" json:"label"`
	URL        string        `yaml:"url" json:"url"`
	MediaTypes slice[string] `yaml:"mediaTypes" json:"mediaTypes"`
}

type User struct {
	Token   string        `yaml:"token" json:"token"`
	Servers slice[Server] `yaml:"servers" json:"servers"`
}

func (u *User) SetToken(token string) error {
	return encryptAndSet(token, func(encrypted string) {
		u.Token = encrypted
	})
}

func (u *User) GetToken() (string, error) {
	return decrypt(u.Token)
}

type Server struct {
	Name                 string        `yaml:"name" json:"name"`
	ListenForUser        string        `yaml:"listenForUser" json:"listenForUser"`
	BlacklistedLibraries slice[string] `yaml:"blacklistedLibraries" json:"blacklistedLibraries"`
	WhitelistedLibraries slice[string] `yaml:"whitelistedLibraries" json:"whitelistedLibraries"`
	IPCPipeNumber        int           `yaml:"ipcPipeNumber" json:"ipcPipeNumber"`
}

func encryptAndSet(decrypted string, set func(string)) error {
	encryptedBytes, err := crypto.AesGcmEncrypt([]byte(decrypted))
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}
	set(encryptionPrefix + base64.StdEncoding.EncodeToString(encryptedBytes))
	return nil
}

func decrypt(encrypted string) (string, error) {
	encryptedBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(encrypted, encryptionPrefix))
	if err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	decrypted, err := crypto.AesGcmDecrypt(encryptedBytes)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(decrypted), nil
}
