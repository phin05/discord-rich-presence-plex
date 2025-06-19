package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Service interface {
	Load() error
	Get(key string) string
	Set(key string, value string) error
}

type service struct {
	filePath string
	data     map[string]string
	mu       sync.RWMutex
}

func NewService(filePath string) Service {
	return &service{
		filePath: filePath,
		data:     map[string]string{},
	}
}

func (s *service) Load() error {
	dataBytes, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return s.save()
		}
		return fmt.Errorf("read file: %w", err)
	}
	if err = json.Unmarshal(dataBytes, &s.data); err != nil {
		return fmt.Errorf("unmarshal data: %w", err)
	}
	return s.save()
}

func (s *service) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key]
}

func (s *service) Set(key string, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return s.save()
}

func (s *service) save() error {
	dataBytes, err := json.Marshal(s.data)
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}
	if err := os.WriteFile(s.filePath, dataBytes, 0600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}
