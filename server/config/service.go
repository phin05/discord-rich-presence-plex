package config

import (
	"bytes"
	"drpp/server/api"
	"drpp/server/fileutil"
	"drpp/server/logger"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

const defaultExtension = ".yml"

var extensions = []string{".yml", ".yaml", ".json"}

type Service struct {
	filePath            string
	useYaml             bool
	config              Config
	mu                  sync.RWMutex
	configChangeHandler func()
}

func NewService(filePath string) (*Service, error) {
	finalFilePath := filePath
	extension := filepath.Ext(finalFilePath)
	if extension == "" {
		for _, ext := range extensions {
			if _, err := os.Stat(filePath + ext); err == nil {
				extension = ext
				break
			}
		}
		if extension == "" {
			extension = defaultExtension
		}
		finalFilePath = filePath + extension
	} else if !slices.Contains(extensions, extension) {
		return nil, fmt.Errorf("unsupported config file extension %q, supported extensions are: %s", extension, strings.Join(extensions, ", "))
	}
	s := &Service{
		filePath: finalFilePath,
		useYaml:  extension != ".json",
		config:   newDefaultConfig(),
	}
	if err := s.load(); err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}
	if err := s.save(); err != nil {
		return nil, fmt.Errorf("save: %w", err)
	}
	if errs := s.config.validate(); len(errs) != 0 {
		return nil, fmt.Errorf("invalid fields (%d):\n%s", len(errs), strings.Join(errs, "\n"))
	}
	return s, nil
}

func (s *Service) load() error {
	configBytes, err := os.ReadFile(s.filePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read file: %w", err)
	}
	configBytes = bytes.TrimSpace(configBytes)
	if len(configBytes) == 0 || string(configBytes) == "null" || string(configBytes) == "{}" {
		return nil
	}
	s.config.Version = 0
	if err := s.unmarshal(configBytes, &s.config); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	if err := s.migrate(configBytes); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

func (s *Service) unmarshal(configBytes []byte, target any) error {
	if s.useYaml {
		return yaml.Unmarshal(configBytes, target)
	}
	return json.Unmarshal(configBytes, target)
}

func (s *Service) migrate(configBytes []byte) error {
	if s.config.Version > currentVersion || s.config.Version < 0 {
		return fmt.Errorf("invalid config version: v%d", s.config.Version)
	}
	if s.config.Version < currentVersion {
		logger.Info("Migrating config from v%d to v%d", s.config.Version, currentVersion)
		extension := filepath.Ext(s.filePath)
		backupFilePath := fmt.Sprintf("%s-v%d%s", strings.TrimSuffix(s.filePath, extension), s.config.Version, extension)
		backupBytes, err := os.ReadFile(s.filePath)
		if err != nil {
			return fmt.Errorf("read existing file: %w", err)
		}
		if err := fileutil.AtomicWrite(backupFilePath, 0o600, func(file *os.File) error {
			_, err := file.Write(backupBytes)
			return err
		}); err != nil {
			return fmt.Errorf("write backup file atomically: %w", err)
		}
		logger.Info("Old config copied to %s", backupFilePath)
		if s.config.Version < 2 {
			if s.config.Version < 1 {
				// v0 to v1
				var oldConfig configV0
				if err := s.unmarshal(configBytes, &oldConfig); err != nil {
					return fmt.Errorf("unmarshal: %w", err)
				}
				if oldConfig.Display.Posters.ImgurClientId != "" {
					if oldConfig.Display.Posters.Enabled {
						s.config.Images.Uploaders.Imgur.Enabled = true
						s.config.Images.Uploaders.Litterbox.Enabled = false
					}
					s.config.Images.Uploaders.Imgur.ClientId.Set(oldConfig.Display.Posters.ImgurClientId)
				}
				s.config.Images.FitInSquare = oldConfig.Display.Posters.Fit
				s.config.Images.MaxSize = oldConfig.Display.Posters.MaxSize
				for _, userV0 := range oldConfig.Users {
					var servers []Server
					for _, serverV0 := range userV0.Servers {
						servers = append(servers, Server{
							Enabled:               true,
							Name:                  serverV0.Name,
							ListenForUser:         serverV0.ListenForUser,
							BlacklistedLibraries:  serverV0.BlacklistedLibraries,
							WhitelistedLibraries:  serverV0.WhitelistedLibraries,
							RequestTimeoutSeconds: 10,
							RetryIntervalSeconds:  5,
							MaxRetriesBeforeExit:  -1,
						})
					}
					user := User{Enabled: true, Servers: servers}
					user.Token.Set(userV0.Token)
					s.config.Plex.Users = append(s.config.Plex.Users, user)
				}
			}
			// For the next version increment - v1 to v2 migration logic goes here
		}
		s.config.Version = currentVersion
		logger.Info("Migration successful")
	}
	return nil
}

func (s *Service) save() error {
	if err := fileutil.AtomicWrite(s.filePath, 0o600, func(file *os.File) error {
		if s.useYaml {
			encoder := yaml.NewEncoder(file)
			encoder.SetIndent(2)
			if err := encoder.Encode(s.config); err != nil {
				return fmt.Errorf("encode: %w", err)
			}
			if err := encoder.Close(); err != nil {
				return fmt.Errorf("close encoder: %w", err)
			}
			return nil
		}
		return json.MarshalWrite(file, s.config, jsontext.EscapeForHTML(false), jsontext.WithIndent("\t"))
	}); err != nil {
		return fmt.Errorf("write file atomically: %w", err)
	}
	return nil
}

func (s *Service) SetConfig(config Config) error {
	config.Version = currentVersion
	if errs := config.validate(); len(errs) != 0 {
		return api.ErrBadRequest(fmt.Sprintf("Invalid input fields (%d)", len(errs)), errs)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
	if err := s.save(); err != nil {
		return fmt.Errorf("save: %w", err)
	}
	if s.configChangeHandler != nil {
		s.configChangeHandler()
	}
	return nil
}

func (s *Service) Config() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *Service) SetConfigChangeHandler(configChangeHandler func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configChangeHandler = configChangeHandler
}
