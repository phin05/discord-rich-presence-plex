package main

import (
	"drpp/internal/cache"
	"drpp/internal/config"
	"drpp/internal/logger"
	"embed"
	"fmt"
	"os"
)

//go:embed all:web/dist
var webFS embed.FS

func main() {
	logger.Info("%s - v%s", config.Name, config.Version)
	if config.IsInContainer {
		ContainerRoutine()
	}
	if err := os.MkdirAll(config.DataDirectoryPath, 0600); err != nil {
		logger.Fatal(err, "Failed to create data directory")
	}
	cacheService := cache.NewCacheService(config.CacheFilePath)
	if err := cacheService.Load(); err != nil {
		logger.Fatal(err, "Failed to load cache")
	}
	configService := config.NewConfigService(config.ConfigFilePathBase)
	if err := configService.Load(); err != nil {
		logger.Fatal(err, "Failed to load config")
	}
	bindAddressPort := fmt.Sprintf("%s:%d", configService.Data.Web.BindAddress, configService.Data.Web.BindPort)
	logger.Info("Web UI Address: http://%s", bindAddressPort)
}
