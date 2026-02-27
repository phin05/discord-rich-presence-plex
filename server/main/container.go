//go:build linux

package main

import (
	"drpp/server/discord"
	"drpp/server/logger"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

const containerCwd = "/app"

var (
	uid               = getEnvInt("DRPP_UID", -1)
	gid               = getEnvInt("DRPP_GID", -1)
	noRuntimeDirChown = getEnvBool("DRPP_NO_RUNTIME_DIR_CHOWN", false)
)

func setupContainer() {
	info, err := os.Stat(discord.RuntimeDirectoryPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			logger.Fatal(nil, "Runtime directory does not exist. Ensure that it is mounted into the container at %q.", discord.RuntimeDirectoryPath)
		}
		logger.Fatal(err, "Failed to stat runtime directory %q", discord.RuntimeDirectoryPath)
	}
	if !info.IsDir() {
		logger.Fatal(nil, "Runtime directory path %q is not a directory", discord.RuntimeDirectoryPath)
	}
	if os.Geteuid() != 0 {
		logger.Warning("Not running as the superuser. Manually ensure appropriate ownership of mounted contents.")
		return
	}
	uid, gid := uid, gid
	var owner string
	if uid == -1 || gid == -1 {
		logger.Warning("Environment variable(s) DRPP_UID and/or DRPP_GID are/is not set. Manually ensure appropriate ownership of the runtime directory.")
		info, err := os.Stat(discord.RuntimeDirectoryPath)
		if err != nil {
			logger.Fatal(err, "Failed to get runtime directory stat info")
		}
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			logger.Fatal(nil, "Failed to get UID/GID from runtime directory stat info")
		}
		uid = int(stat.Uid)
		gid = int(stat.Gid)
		owner = fmt.Sprintf("%d:%d", uid, gid)
	} else {
		owner = fmt.Sprintf("%d:%d", uid, gid)
		if noRuntimeDirChown {
			logger.Warning("Environment variable DRPP_NO_RUNTIME_DIR_CHOWN is set to true. Manually ensure appropriate ownership of the runtime directory.")
		} else {
			if err := os.Chmod(discord.RuntimeDirectoryPath, 0o700); err != nil { //nolint:gosec
				logger.Warning("Failed to chmod runtime directory: %v", err)
			}
			cmd := exec.Command("chown", "-R", owner, discord.RuntimeDirectoryPath) //nolint:gosec
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				logger.Warning("Failed to chown runtime directory: %v", err)
			}
		}
	}
	cmd := exec.Command("chown", "-R", owner, containerCwd) //nolint:gosec
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logger.Warning("Failed to chown %q: %v", containerCwd, err)
	}
	if err := syscall.Setgid(gid); err != nil {
		logger.Fatal(err, "Failed to set GID")
	}
	if err := syscall.Setuid(uid); err != nil {
		logger.Fatal(err, "Failed to set UID")
	}
}

func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseBool(val); err == nil {
			return parsed
		}
	}
	return defaultValue
}
