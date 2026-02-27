package cache

import (
	"bytes"
	"drpp/server/fileutil"
	"drpp/server/logger"
	"encoding/json" // TODO: Use encoding/json/v2
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"
	"time"
)

const cleanerInterval = 5 * time.Minute

type Service struct {
	filePath string
	cache    map[string]entry
	mu       sync.RWMutex
}

type entry struct {
	Value  string `json:"value"`
	Expiry int64  `json:"expiry"`
}

func NewService(filePath string) (*Service, error) {
	s := &Service{
		filePath: filePath,
	}
	if err := s.load(); err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}
	if s.cache == nil {
		s.cache = make(map[string]entry)
	}
	if err := s.save(); err != nil {
		return nil, fmt.Errorf("save: %w", err)
	}
	// TODO: Graceful shutdown
	// Cache service is used as a singleton and never stopped/disposed in this app, so this is safe for now
	go s.cleaner()
	return s, nil
}

func (s *Service) load() error {
	cacheBytes, err := os.ReadFile(s.filePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read file: %w", err)
	}
	cacheBytes = bytes.TrimSpace(cacheBytes)
	if len(cacheBytes) == 0 || string(cacheBytes) == "null" || string(cacheBytes) == "{}" {
		return nil
	}
	if err := json.Unmarshal(cacheBytes, &s.cache); err != nil {
		// Intentionally ignore type errors
		if _, ok := errors.AsType[*json.UnmarshalTypeError](err); !ok {
			return fmt.Errorf("unmarshal: %w", err)
		}
	}
	return nil
}

func (s *Service) save() error {
	if err := fileutil.AtomicWrite(s.filePath, 0o600, func(file *os.File) error {
		encoder := json.NewEncoder(file)
		encoder.SetEscapeHTML(false)
		return encoder.Encode(s.cache)
	}); err != nil {
		return fmt.Errorf("write file atomically: %w", err)
	}
	return nil
}

func (s *Service) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, found := s.cache[key]
	// Expiry <=0 means it never expires
	if !found || (entry.Expiry > 0 && time.Now().Unix() > entry.Expiry) {
		return ""
	}
	return entry.Value
}

func (s *Service) Set(key string, value string, lifespan time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var expiry int64
	// Lifespan/expiry <=0 means it never expires
	if lifespan > 0 {
		expiry = time.Now().Unix() + int64(lifespan.Seconds())
	} else {
		expiry = 0
	}
	s.cache[key] = entry{Value: value, Expiry: expiry}
	if err := s.save(); err != nil {
		return fmt.Errorf("save: %w", err)
	}
	return nil
}

func (s *Service) cleaner() {
	ticker := time.NewTicker(cleanerInterval)
	defer ticker.Stop()
	for range ticker.C {
		s.clean()
	}
}

func (s *Service) clean() {
	s.mu.Lock()
	defer s.mu.Unlock()
	var changed bool
	now := time.Now().Unix()
	for key, entry := range s.cache {
		if entry.Value == "" || (entry.Expiry > 0 && now > entry.Expiry) {
			// In Go, it's safe to delete keys from a map while iterating over it
			delete(s.cache, key)
			changed = true
		}
	}
	if !changed {
		return
	}
	if err := s.save(); err != nil {
		logger.Error(err, "Failed to save cache after removing empty and expired entries")
	}
}
