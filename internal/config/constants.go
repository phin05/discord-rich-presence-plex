package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

const (
	Name                     = "Discord Rich Presence for Plex"
	Version                  = "3.0.0"
	PlexClientID             = "discord-rich-presence-plex"
	DiscordClientID   uint64 = 413407336082833418
	DataDirectoryPath        = "data"
	ContainerCwdPath         = "/app"
	IsUnix                   = runtime.GOOS == "linux" || runtime.GOOS == "darwin"
)

var (
	ConfigFilePathBase   = filepath.Join(DataDirectoryPath, "config")
	CacheFilePath        = filepath.Join(DataDirectoryPath, "cache.json")
	ProcessID            = os.Getpid()
	IsInContainer        = os.Getenv("DRPP_IS_IN_CONTAINER") == "true"
	RuntimeDirectoryPath = runtimeDirectoryPath()
	IpcPipeBase          = ipcPipeBase()
	UID                  = getEnvInt("DRPP_UID", -1)
	GID                  = getEnvInt("DRPP_GID", -1)
	NoRuntimeDirChown    = os.Getenv("DRPP_NO_RUNTIME_DIR_CHOWN") == "true"
)

func runtimeDirectoryPath() string {
	if IsInContainer {
		return "/run/app"
	}
	for _, key := range []string{"XDG_RUNTIME_DIR", "TMPDIR", "TMP", "TEMP"} {
		if path := os.Getenv(key); path != "" {
			return path
		}
	}
	return "/tmp"
}

func ipcPipeBase() string {
	if IsUnix {
		return RuntimeDirectoryPath
	}
	return `\\?\pipe`
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return i
}
