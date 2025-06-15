package config

import (
	"drpp/internal/logger"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

type configService struct {
	mu       sync.Mutex
	filePath string
	useYaml  bool
	Data     Config
}

var supportedExts = []string{"json", "yml", "yaml"}

func NewConfigService(filePathBase string) *configService {
	var filePath string
	var ext string
	for _, ext = range supportedExts {
		filePath = fmt.Sprintf("%s.%s", filePathBase, ext)
		info, err := os.Stat(filePath)
		if err == nil && !info.IsDir() {
			break
		}
	}
	return &configService{
		filePath: filePath,
		useYaml:  ext != "json",
		Data: Config{
			Web: Web{
				BindAddress: "127.0.0.1",
				BindPort:    8040,
			},
		},
	}
}

func (c *configService) Load() error {
	configData, err := os.ReadFile(c.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return c.Save()
		}
		return fmt.Errorf("read config file: %w", err)
	}
	if c.useYaml {
		err = yaml.Unmarshal(configData, &c.Data)
	} else {
		err = json.Unmarshal(configData, &c.Data)
	}
	if err != nil {
		return fmt.Errorf("unmarshal config data: %w", err)
	}
	return c.Save()
}

func (c *configService) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.autofillAndValidate(); err != nil {
		return fmt.Errorf("autofill and validate: %w", err)
	}
	file, err := os.Create(c.filePath)
	if err != nil {
		return fmt.Errorf("open config file: %w", err)
	}
	defer file.Close()
	if c.useYaml {
		encoder := yaml.NewEncoder(file)
		defer encoder.Close()
		encoder.SetIndent(2)
		err = encoder.Encode(&c.Data)
	} else {
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "\t")
		err = encoder.Encode(&c.Data)
	}
	if err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}

func (c *configService) autofillAndValidate() error {
	for _, user := range c.Data.Users {
		if !strings.HasPrefix(user.Token, encryptionPrefix) {
			if err := user.SetToken(user.Token); err != nil {
				return err
			}
		}
	}
	switch c.Data.Display.ProgressMode {
	case "off", "elapsed", "remaining", "bar":
	default:
		logger.Warning("Invalid 'display.progressMode' value '%s'. Resetting to default value 'bar'.", c.Data.Display.ProgressMode) // TODO: Avoid using logger here?
		c.Data.Display.ProgressMode = "bar"
	}
	return nil
}
