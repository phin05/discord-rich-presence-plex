package config

import (
	"drpp/internal/crypto"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/go-playground/validator/v10"
)

func newDefaultConfig() *Config {
	return &Config{
		DiscordClientID: DiscordClientID,
		Web: web{
			BindAddress: "127.0.0.1",
			BindPort:    8040,
		},
		Display: display{
			Genres:       true,
			Album:        true,
			AlbumImage:   true,
			Artist:       true,
			ArtistImage:  true,
			Year:         true,
			ProgressMode: "bar",
			Posters: posters{
				Fit: true,
			},
		},
	}
}

var validate = validator.New()

func (c *Config) validate() []string {
	errs := []string{}
	err := validate.Struct(c)
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrs {
			var separator string
			if fieldErr.Param() != "" {
				separator = ":"
			}
			errs = append(errs, fmt.Sprintf("%s=%v (%s%s%s)", fieldErr.Namespace(), fieldErr.Value(), fieldErr.Tag(), separator, fieldErr.Param()))
		}
	} else if err != nil {
		errs = append(errs, err.Error())
	}
	return errs
}

func (c Config) deepCopy() *Config {
	c.Users = slices.Clone(c.Users)
	for i := range c.Users {
		user := &c.Users[i]
		user.Servers = slices.Clone(user.Servers)
		for j := range user.Servers {
			server := &user.Servers[j]
			server.BlacklistedLibraries = slices.Clone(server.BlacklistedLibraries)
			server.WhitelistedLibraries = slices.Clone(server.WhitelistedLibraries)
		}
	}
	c.Display.Buttons = slices.Clone(c.Display.Buttons)
	for i := range c.Display.Buttons {
		button := &c.Display.Buttons[i]
		button.MediaTypes = slices.Clone(button.MediaTypes)
	}
	return &c
}

type slice[T any] []T

func (s slice[T]) MarshalJSON() ([]byte, error) {
	if s == nil {
		return json.Marshal([]T{})
	}
	return json.Marshal([]T(s))
}

type Config struct {
	DiscordClientID uint64      `yaml:"discordClientID" json:"discordClientID" validate:"required"`
	Web             web         `yaml:"web" json:"web"`
	Display         display     `yaml:"display" json:"display"`
	Users           slice[user] `yaml:"users" json:"users" validate:"dive"`
}

type web struct {
	BindAddress string `yaml:"bindAddress" json:"bindAddress" validate:"ip"`
	BindPort    int    `yaml:"bindPort" json:"bindPort" validate:"min=1,max=65535"`
}

type display struct {
	Duration     bool          `yaml:"duration" json:"duration"`
	Genres       bool          `yaml:"genres" json:"genres"`
	Album        bool          `yaml:"album" json:"album"`
	AlbumImage   bool          `yaml:"albumImage" json:"albumImage"`
	Artist       bool          `yaml:"artist" json:"artist"`
	ArtistImage  bool          `yaml:"artistImage" json:"artistImage"`
	Year         bool          `yaml:"year" json:"year"`
	StatusIcon   bool          `yaml:"statusIcon" json:"statusIcon"`
	ProgressMode string        `yaml:"progressMode" json:"progressMode" validate:"oneof='off' 'elapsed' 'remaining' 'bar'"`
	Paused       bool          `yaml:"paused" json:"paused"`
	Posters      posters       `yaml:"posters" json:"posters"`
	Buttons      slice[button] `yaml:"buttons" json:"buttons" validate:"dive"`
}

type posters struct {
	Enabled       bool          `yaml:"enabled" json:"enabled"`
	ImgurClientID crypto.String `yaml:"imgurClientID" json:"imgurClientID"`
	MaxSize       int           `yaml:"maxSize" json:"maxSize"`
	Fit           bool          `yaml:"fit" json:"fit"`
}

type button struct {
	Label      string        `yaml:"label" json:"label" validate:"required"`
	URL        string        `yaml:"url" json:"url" validate:"required"`
	MediaTypes slice[string] `yaml:"mediaTypes" json:"mediaTypes" validate:"unique,dive,oneof='movie' 'episode' 'live_episode' 'track' 'clip'"`
}

type user struct {
	Token   crypto.String `yaml:"token" json:"token"`
	Servers slice[server] `yaml:"servers" json:"servers" validate:"dive"`
}

type server struct {
	Name                 string        `yaml:"name" json:"name" validate:"required"`
	ListenForUser        string        `yaml:"listenForUser" json:"listenForUser"`
	BlacklistedLibraries slice[string] `yaml:"blacklistedLibraries" json:"blacklistedLibraries" validate:"unique"`
	WhitelistedLibraries slice[string] `yaml:"whitelistedLibraries" json:"whitelistedLibraries" validate:"unique"`
	IPCPipeNumber        int           `yaml:"ipcPipeNumber" json:"ipcPipeNumber" validate:"min=-1,max=9"`
}
