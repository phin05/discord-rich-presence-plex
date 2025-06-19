package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

type service struct {
	filePath string
	useYaml  bool
	config   *Config
	mu       sync.Mutex
}

var exts = []string{"yml", "yaml", "json"}

func NewService(filePathBase string) *service {
	var filePath string
	var ext string
	for _, ext = range exts {
		filePath = filePathBase + "." + ext
		info, err := os.Stat(filePath)
		if err == nil && !info.IsDir() {
			break
		}
	}
	return &service{
		filePath: filePath,
		useYaml:  ext != "json",
		config:   newDefaultConfig(),
	}
}

func (s *service) Load() error {
	configBytes, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return s.Save()
		}
		return fmt.Errorf("read file: %w", err)
	}
	if s.useYaml {
		err = yaml.Unmarshal(configBytes, &s.config)
	} else {
		err = json.Unmarshal(configBytes, &s.config)
	}
	if err != nil {
		return fmt.Errorf("unmarshal data: %w", err)
	}
	if errs := s.config.validate(); len(errs) != 0 {
		return fmt.Errorf("invalid fields:\n%s", strings.Join(errs, "\n"))
	}
	return s.Save()
}

func (s *service) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := os.Create(s.filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()
	if s.useYaml {
		encoder := yaml.NewEncoder(file)
		defer encoder.Close()
		encoder.SetIndent(2)
		err = encoder.Encode(s.config)
	} else {
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "\t")
		err = encoder.Encode(s.config)
	}
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}
	return nil
}

func (s *service) Config() Config {
	return *s.config
}
