package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	logFileFlag := stringFlagWithEnv("log-file", "", "Path to log file")
	dataDirFlag := stringFlagWithEnv("data-dir", "", `Path to data directory. Defaults to "data" relative to executable.`)
	configFileFlag := stringFlagWithEnv("config-file", "", `Path to config file. Defaults to "config.{yml,yaml,json}" inside data directory.`)
	cacheFileFlag := stringFlagWithEnv("cache-file", "", `Path to cache file. Defaults to "cache.json" inside data directory.`)
	disableWebUiFlag := boolFlagWithEnv("disable-web-ui", false, "Disable web interface")
	disableWebUiLaunchFlag := boolFlagWithEnv("disable-web-ui-launch", false, "Disable launching web interface on startup")
	disableSystrayFlag := boolFlagWithEnv("disable-systray", false, "Disable system tray icon")
	flag.Parse()
	logFilePath = *logFileFlag
	dataDirPath = *dataDirFlag
	configFilePath = *configFileFlag
	cacheFilePath = *cacheFileFlag
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
	return "DRPP_" + strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
}
