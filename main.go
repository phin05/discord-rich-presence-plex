package main

import (
	"drpp/internal/cache"
	"drpp/internal/config"
	"drpp/internal/exc"
	"drpp/internal/logger"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
)

//go:embed all:web/dist
var webFS embed.FS

type httpResponseWriter struct {
	http.ResponseWriter
}

func main() {
	logger.Info("%s - v%s", config.Name, config.Version)
	ContainerRoutine()
	if err := os.MkdirAll(config.DataDirectoryPath, 0600); err != nil {
		logger.Fatal(err, "Failed to create data directory")
	}
	cacheService := cache.NewService(config.CacheFilePath)
	if err := cacheService.Load(); err != nil {
		logger.Fatal(err, "Failed to load cache")
	}
	configService := config.NewService(config.ConfigFilePathBase)
	if err := configService.Load(); err != nil {
		logger.Fatal(err, "Failed to load config")
	}
	mux := http.NewServeMux()
	webFSSubbed, err := fs.Sub(webFS, "web/dist")
	if err != nil {
		logger.Fatal(err, "Failed to get embedded web dist subtree")
	}
	fileServer := http.FileServer(http.FS(webFSSubbed))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fileServer.ServeHTTP(&httpResponseWriter{w}, r)
	})
	configController := config.NewController(configService)
	mux.HandleFunc("GET /api/config", exc.WithErrorHandling(configController.Get))
	mux.HandleFunc("PUT /api/config", exc.WithErrorHandling(configController.Put))
	mux.HandleFunc("PATCH /api/config", exc.WithErrorHandling(configController.Patch))
	webAddr := fmt.Sprintf("%s:%d", configService.Config().Web.BindAddress, configService.Config().Web.BindPort)
	logger.Info("Web UI Address: http://%s", webAddr)
	if err := http.ListenAndServe(webAddr, exc.WithPanicRecovery(mux)); err != nil {
		logger.Fatal(err, "Failed to run HTTP server")
	}
}
