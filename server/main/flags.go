package main

import (
	"drpp/server/config"
	"drpp/server/logger"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	dataDirName         = "drpp"
	relativeDataDirName = "data"
	envVarPrefix        = "DRPP_"
)

var (
	logFilePath        string
	dataDirPath        string
	configFilePath     string
	cacheFilePath      string
	disableWebUi       bool
	disableWebUiLaunch bool
	disableSystray     bool
)

func parseFlags() {
	var defaultDataDirPath string
	if config.Containerised {
		defaultDataDirPath = filepath.Join(containerCwd, relativeDataDirName)
	} else {
		exe, err := os.Executable()
		if err != nil {
			logger.Fatal(err, "Failed to get executable path")
		}
		legacyDataDirPath := filepath.Join(filepath.Dir(exe), relativeDataDirName)
		if stat, err := os.Stat(legacyDataDirPath); err == nil && stat.IsDir() {
			defaultDataDirPath = legacyDataDirPath
		} else {
			userConfigDir, err := os.UserConfigDir()
			if err != nil {
				logger.Fatal(err, "Failed to get user config directory")
			}
			defaultDataDirPath = filepath.Join(userConfigDir, dataDirName)
		}
	}
	logFileFlag := stringFlagWithEnv("log-file", "", "Path to log file. Disabled if empty.")
	dataDirFlag := stringFlagWithEnv("data-dir", defaultDataDirPath, `Path to data directory`)
	configFileFlag := stringFlagWithEnv("config-file", "", `Path to config file. Defaults to "config.yml" inside data directory.`)
	cacheFileFlag := stringFlagWithEnv("cache-file", "", `Path to cache file. Defaults to "cache.json" inside data directory.`)
	disableWebUiFlag := boolFlagWithEnv("disable-web-ui", false, "Disable web interface")
	disableWebUiLaunchFlag := boolFlagWithEnv("disable-web-ui-launch", false, "Disable launching web interface on startup (override flag meant for autostart)")
	disableSystrayFlag := boolFlagWithEnv("disable-systray", false, "Disable system tray icon")
	flag.Parse()
	logFilePath = *logFileFlag
	dataDirPath = *dataDirFlag
	configFilePath = *configFileFlag
	if configFilePath == "" {
		configFilePath = filepath.Join(dataDirPath, "config")
	}
	cacheFilePath = *cacheFileFlag
	if cacheFilePath == "" {
		cacheFilePath = filepath.Join(dataDirPath, "cache.json")
	}
	disableWebUi = *disableWebUiFlag
	disableWebUiLaunch = *disableWebUiLaunchFlag
	disableSystray = *disableSystrayFlag
}

func stringFlagWithEnv(name string, value string, usage string) *string {
	key := getEnvKey(name)
	if val := os.Getenv(key); val != "" {
		value = val
	}
	return flag.String(name, value, fmt.Sprintf("%s (env: %s)", usage, key))
}

func boolFlagWithEnv(name string, value bool, usage string) *bool {
	key := getEnvKey(name)
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseBool(val); err == nil {
			value = parsed
		}
	}
	return flag.Bool(name, value, fmt.Sprintf("%s (env: %s)", usage, key))
}

func getEnvKey(name string) string {
	return envVarPrefix + strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
}
