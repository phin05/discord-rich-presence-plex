package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type cacheService struct {
	mu       sync.RWMutex
	filePath string
	data     map[string]string
}

func NewCacheService(filePath string) *cacheService {
	return &cacheService{
		filePath: filePath,
		data:     map[string]string{},
	}
}

func (c *cacheService) Load() error {
	cacheData, err := os.ReadFile(c.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return c.save()
		}
		return fmt.Errorf("read cache file: %w", err)
	}
	if err = json.Unmarshal(cacheData, &c.data); err != nil {
		return fmt.Errorf("unmarshal cache data: %w", err)
	}
	return c.save()
}

func (c *cacheService) Get(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data[key]
}

func (c *cacheService) Set(key string, value string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
	return c.save()
}

func (c *cacheService) save() error {
	cacheData, err := json.Marshal(c.data)
	if err != nil {
		return fmt.Errorf("marshal cache data: %w", err)
	}
	if err := os.WriteFile(c.filePath, cacheData, 0600); err != nil {
		return fmt.Errorf("write cache file: %w", err)
	}
	return nil
}
