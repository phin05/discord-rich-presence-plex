package config

import (
	"drpp/server/crypto"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

const currentVersion = 1

type Config struct {
	Version int     `yaml:"version" json:"version"`
	Web     Web     `yaml:"web" json:"web"`
	Logger  Logger  `yaml:"logger" json:"logger"`
	Discord Discord `yaml:"discord" json:"discord"`
	Images  Images  `yaml:"images" json:"images"`
	Plex    Plex    `yaml:"plex" json:"plex"`
}

type Web struct {
	LaunchOnStartup bool     `yaml:"launchOnStartup" json:"launchOnStartup"`
	BindAddress     string   `yaml:"bindAddress" json:"bindAddress" validate:"ip"`
	BindPort        int      `yaml:"bindPort" json:"bindPort" validate:"min=1,max=65535"`
	AllowedNetworks []string `yaml:"allowedNetworks" json:"allowedNetworks" validate:"unique,dive,cidr"`
	TrustedProxies  []string `yaml:"trustedProxies" json:"trustedProxies" validate:"unique,dive,cidr"`
}

type Logger struct {
	EnableDebugOutput bool `yaml:"enableDebugOutput" json:"enableDebugOutput"`
}

type Discord struct {
	ClientId           string       `yaml:"clientId" json:"clientId" validate:"required"`
	IpcPipeNumber      int          `yaml:"ipcPipeNumber" json:"ipcPipeNumber" validate:"min=-1,max=9"`
	IpcTimeoutSeconds  int          `yaml:"ipcTimeoutSeconds" json:"ipcTimeoutSeconds" validate:"min=1"`
	RateLimit          int          `yaml:"rateLimit" json:"rateLimit" validate:"min=1"`
	StopTimeoutSeconds int          `yaml:"stopTimeoutSeconds" json:"stopTimeoutSeconds" validate:"min=0"`
	IdleTimeoutSeconds int          `yaml:"idleTimeoutSeconds" json:"idleTimeoutSeconds" validate:"min=1"`
	DisplayRules       DisplayRules `yaml:"displayRules" json:"displayRules"`
}

type DisplayRules struct {
	Movie       DisplayRule `yaml:"movie" json:"movie"`
	Episode     DisplayRule `yaml:"episode" json:"episode"`
	Track       DisplayRule `yaml:"track" json:"track"`
	Clip        DisplayRule `yaml:"clip" json:"clip"`
	LiveEpisode DisplayRule `yaml:"liveEpisode" json:"liveEpisode"`
}

type DisplayRule struct {
	Details             string   `yaml:"details" json:"details"`
	State               string   `yaml:"state" json:"state"`
	StatusType          string   `yaml:"statusType" json:"statusType" validate:"required"`
	LargeImage          string   `yaml:"largeImage" json:"largeImage"`
	LargeText           string   `yaml:"largeText" json:"largeText"`
	SmallImage          string   `yaml:"smallImage" json:"smallImage"`
	SmallText           string   `yaml:"smallText" json:"smallText"`
	DetailsUrl          string   `yaml:"detailsUrl" json:"detailsUrl"`
	StateUrl            string   `yaml:"stateUrl" json:"stateUrl"`
	LargeUrl            string   `yaml:"largeUrl" json:"largeUrl"`
	SmallUrl            string   `yaml:"smallUrl" json:"smallUrl"`
	ProgressMode        string   `yaml:"progressMode" json:"progressMode" validate:"required"`
	PauseTimeoutSeconds int      `yaml:"pauseTimeoutSeconds" json:"pauseTimeoutSeconds" validate:"min=-1"`
	Buttons             []Button `yaml:"buttons" json:"buttons" validate:"dive"`
}

type Button struct {
	Label string `yaml:"label" json:"label" validate:"required"`
	Url   string `yaml:"url" json:"url" validate:"required"`
}

type Images struct {
	FitInSquare          bool      `yaml:"fitInSquare" json:"fitInSquare"`
	MaxSize              int       `yaml:"maxSize" json:"maxSize" validate:"min=1"`
	UploadTimeoutSeconds int       `yaml:"uploadTimeoutSeconds" json:"uploadTimeoutSeconds" validate:"min=1"`
	Uploaders            Uploaders `yaml:"uploaders" json:"uploaders"`
}

type Uploaders struct {
	Litterbox Litterbox `yaml:"litterbox" json:"litterbox"`
	ImgBb     ImgBb     `yaml:"imgBb" json:"imgBb"`
	Imgur     Imgur     `yaml:"imgur" json:"imgur"`
	Copyparty Copyparty `yaml:"copyparty" json:"copyparty"`
}

type Litterbox struct {
	Enabled     bool `yaml:"enabled" json:"enabled"`
	ExpiryHours int  `yaml:"expiryHours" json:"expiryHours" validate:"oneof=1 12 24 72"`
}

type ImgBb struct {
	Enabled       bool          `yaml:"enabled" json:"enabled"`
	ApiKey        crypto.String `yaml:"apiKey" json:"apiKey" validate:"required_if=Enabled true"`
	ExpiryMinutes int           `yaml:"expiryMinutes" json:"expiryMinutes" validate:"min=0,max=259200"`
}

type Imgur struct {
	Enabled  bool          `yaml:"enabled" json:"enabled"`
	ClientId crypto.String `yaml:"clientId" json:"clientId" validate:"required_if=Enabled true"`
}

type Copyparty struct {
	Enabled       bool          `yaml:"enabled" json:"enabled"`
	Url           string        `yaml:"url" json:"url" validate:"required_if=Enabled true,omitempty,http_url"`
	Password      crypto.String `yaml:"password" json:"password"`
	ExpiryMinutes int           `yaml:"expiryMinutes" json:"expiryMinutes" validate:"min=0"`
}

type Plex struct {
	Users []User `yaml:"users" json:"users" validate:"dive"`
}

type User struct {
	Enabled bool          `yaml:"enabled" json:"enabled"`
	Name    string        `yaml:"name" json:"name"`
	Token   crypto.String `yaml:"token" json:"token" validate:"required"`
	Servers []Server      `yaml:"servers" json:"servers" validate:"dive"`
}

type Server struct {
	Enabled               bool     `yaml:"enabled" json:"enabled"`
	Name                  string   `yaml:"name" json:"name" validate:"required"`
	Url                   string   `yaml:"url" json:"url" validate:"omitempty,http_url"`
	ListenForUser         string   `yaml:"listenForUser" json:"listenForUser"`
	BlacklistedLibraries  []string `yaml:"blacklistedLibraries" json:"blacklistedLibraries" validate:"unique,dive,required"`
	WhitelistedLibraries  []string `yaml:"whitelistedLibraries" json:"whitelistedLibraries" validate:"unique,dive,required"`
	RequestTimeoutSeconds int      `yaml:"requestTimeoutSeconds" json:"requestTimeoutSeconds" validate:"min=1"`
	RetryIntervalSeconds  int      `yaml:"retryIntervalSeconds" json:"retryIntervalSeconds" validate:"min=1"`
	MaxRetriesBeforeExit  int      `yaml:"maxRetriesBeforeExit" json:"maxRetriesBeforeExit" validate:"min=-1"`
}

var validate = validator.New(validator.WithRequiredStructEnabled())

func (c *Config) validate() []string {
	err := validate.Struct(c)
	if err == nil {
		return nil
	}
	if _, ok := errors.AsType[*validator.InvalidValidationError](err); ok {
		return []string{"Root object is invalid"}
	}
	validationErrs, ok := errors.AsType[validator.ValidationErrors](err)
	if !ok {
		return []string{"Validation failed due to an unexpected error"}
	}
	errs := make([]string, len(validationErrs))
	for i, fieldErr := range validationErrs {
		_, namespace, _ := strings.Cut(fieldErr.Namespace(), ".")
		tag := fieldErr.Tag()
		if param := fieldErr.Param(); param != "" {
			tag += "=" + param
		}
		errs[i] = fmt.Sprintf("%s;%s", namespace, tag)
	}
	return errs
}
