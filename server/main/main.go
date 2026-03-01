package main

import (
	"bytes"
	"context"
	"drpp/server/api"
	"drpp/server/cache"
	"drpp/server/config"
	"drpp/server/discord"
	"drpp/server/images/copyparty"
	"drpp/server/images/imgbb"
	"drpp/server/images/imgur"
	"drpp/server/images/litterbox"
	"drpp/server/logger"
	"drpp/server/mediator"
	"drpp/server/plex"
	"drpp/web"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var version = "0.0.0-dev"

const containerCwd = "/app"

func main() {

	if config.Containerised {
		setupContainer()
	}

	parseFlags()

	if logFilePath != "" {
		if err := logger.SetLogFile(logFilePath); err != nil {
			logger.Fatal(err, "Failed to set log file")
		}
	}

	logger.Info("Discord Rich Presence for Plex (DRPP) - v%s", version)
	logger.Info("Data Directory: %s", dataDirPath)

	configService, err := config.NewService(configFilePath)
	if err != nil {
		logger.Fatal(err, "Failed to load config")
	}
	configController := config.NewController(configService)
	cfg := configService.Config()

	cacheService, err := cache.NewService(cacheFilePath)
	if err != nil {
		logger.Fatal(err, "Failed to load cache")
	}

	var fatalExit atomic.Bool
	shutdownCtx, shutdown := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	fatalShutdown := func() {
		fatalExit.Store(true)
		shutdown()
	}

	var server *api.Server
	var mediatorService *mediator.Service
	setup := func(cfg config.Config, setupServer bool) {
		logger.EnableDebugOutput.Store(cfg.Logger.EnableDebugOutput)
		if !disableWebUi && setupServer {
			server = api.NewServer(shutdownCtx, fmt.Sprintf("%s:%d", cfg.Web.BindAddress, cfg.Web.BindPort), web.BuildOutput, devMode, cfg.Web.AllowedNetworks, cfg.Web.TrustedProxies)
			configController.RegisterRoutes(server)
			server.RegisterCustomRoute(logger.SseHandler, http.MethodGet, "logs")
			go func() {
				if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					if !disableWebUiLaunch {
						if errno, ok := errors.AsType[syscall.Errno](err); ok {
							if errno == syscall.EADDRINUSE || (runtime.GOOS == "windows" && errno == 10048) { // 10048 is WSAEADDRINUSE
								webUiAddress := fmt.Sprintf("http://%s:%d", cfg.Web.BindAddress, cfg.Web.BindPort) //nolint:nosprintfhostport
								logger.Info("Address %s already in use, launching URL", webUiAddress)
								launchUrl(webUiAddress)
							}
						}
					}
					logger.Error(err, "Failed to run HTTP server")
					fatalShutdown()
				}
			}()
		}
		discordService := discord.NewService(cfg.Discord.ClientId, cfg.Discord.IpcPipeNumber, cfg.Discord.RateLimit)
		var plexServices []*plex.Service
		for _, user := range cfg.Plex.Users {
			if !user.Enabled {
				continue
			}
			for _, server := range user.Servers {
				if !server.Enabled {
					continue
				}
				plexServices = append(plexServices, plex.NewService(user.Token.Value(), server, fatalShutdown))
			}
		}
		if len(plexServices) == 0 {
			logger.Warning("Plex connection not configured. Add a Plex user and server to finish setting up.")
		}
		var imageService mediator.ImageService
		switch {
		case cfg.Images.Uploaders.Litterbox.Enabled:
			imageService = litterbox.NewService(cfg.Images.Uploaders.Litterbox.ExpiryHours)
		case cfg.Images.Uploaders.ImgBb.Enabled:
			imageService = imgbb.NewService(cfg.Images.Uploaders.ImgBb.ApiKey.Value(), cfg.Images.Uploaders.ImgBb.ExpiryMinutes)
		case cfg.Images.Uploaders.Imgur.Enabled:
			imageService = imgur.NewService(cfg.Images.Uploaders.Imgur.ClientId.Value())
		case cfg.Images.Uploaders.Copyparty.Enabled:
			imageService = copyparty.NewService(cfg.Images.Uploaders.Copyparty.Url, cfg.Images.Uploaders.Copyparty.Password.Value(), cfg.Images.Uploaders.Copyparty.ExpiryMinutes)
		}
		mediatorService = mediator.NewService(
			discordService,
			plexServices,
			cacheService,
			imageService,
			cfg.Images,
			cfg.Discord,
		)
		mediatorService.Start() //nolint:contextcheck
	}
	setup(cfg, true)

	// Allow at most one reload to be queued at a time
	reloadCh := make(chan struct{}, 1)
	var mu sync.Mutex
	lastWebConfig, _ := json.Marshal(cfg.Web)
	reload := func() {
		logger.Info("Reloading application due to config change")
		mu.Lock()
		defer mu.Unlock()
		mediatorService.Stop()
		cfg := configService.Config()
		webConfig, _ := json.Marshal(cfg.Web)
		setupServer := !bytes.Equal(webConfig, lastWebConfig) // TODO: This is a bit hacky
		if server != nil && setupServer {
			if err := server.Stop(); err != nil {
				logger.Error(err, "Failed to gracefully stop HTTP server")
			}
		}
		lastWebConfig = webConfig
		setup(cfg, setupServer)
	}
	go func() {
		for range reloadCh {
			reload()
		}
	}()
	configService.SetConfigChangeHandler(func() {
		select {
		case reloadCh <- struct{}{}:
		default:
		}
	})

	if !config.Containerised {
		time.AfterFunc(time.Second, func() { // Delayed to allow initial setup to complete
			if shutdownCtx.Err() != nil {
				return
			}
			webUiAddress := fmt.Sprintf("http://%s:%d", cfg.Web.BindAddress, cfg.Web.BindPort) //nolint:nosprintfhostport
			if !disableWebUi && cfg.Web.LaunchOnStartup && !disableWebUiLaunch {
				go launchUrl(webUiAddress)
			}
			if !disableSystray {
				var iconBytes []byte
				if runtime.GOOS == "windows" {
					iconBytes, err = fs.ReadFile(web.BuildOutput, "favicon.ico")
				} else {
					iconBytes, err = fs.ReadFile(web.BuildOutput, "favicon.png")
				}
				if err != nil {
					logger.Error(err, "Failed to read icon from web build output")
					fatalShutdown()
				} else {
					go runSystray(webUiAddress, iconBytes, shutdown)
				}
			}
		})
	}

	<-shutdownCtx.Done()
	logger.Info("Received shutdown signal, shutting down gracefully")
	mu.Lock()
	mediatorService.Stop()
	if server != nil {
		if err := server.Stop(); err != nil {
			logger.Error(err, "Failed to stop HTTP server")
			fatalShutdown()
		}
	}
	_ = logger.CloseLogFile()
	logger.Info("Goodbye")
	if fatalExit.Load() {
		os.Exit(1)
	}

}

func launchUrl(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		logger.Error(nil, "Unsupported OS %q for launching URL", runtime.GOOS)
		return
	}
	if err := cmd.Start(); err != nil {
		logger.Error(err, "Failed to launch URL")
	}
	if err := cmd.Wait(); err != nil {
		logger.Error(err, "Failed to wait for URL launch")
	}
}
