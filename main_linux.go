//go:build linux

package main

import (
	"drpp/internal/config"
	"drpp/internal/logger"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func ContainerRoutine() {
	if info, err := os.Stat(config.RuntimeDirectoryPath); err != nil || !info.IsDir() {
		logger.Fatal(err, "Runtime directory does not exist. Ensure that it is mounted into the container at '%s'.", config.RuntimeDirectoryPath)
	}
	if os.Geteuid() != 0 {
		logger.Warning("Not running as the superuser. Manually ensure appropriate ownership of mounted contents.")
		return
	}
	uid, gid := config.UID, config.GID
	if uid == -1 || gid == -1 {
		logger.Warning("Environment variable(s) DRPP_UID and/or DRPP_GID are/is not set. Manually ensure appropriate ownership of the runtime directory.")
		info, err := os.Stat(config.RuntimeDirectoryPath)
		if err != nil {
			logger.Fatal(err, "Failed to get runtime directory stat info")
		}
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			logger.Fatal(nil, "Failed to get UID/GID from runtime directory stat info")
		}
		uid = int(stat.Uid)
		gid = int(stat.Gid)
	} else {
		if config.NoRuntimeDirChown {
			logger.Warning("Environment variable DRPP_NO_RUNTIME_DIR_CHOWN is set to true. Manually ensure appropriate ownership of the runtime directory.")
		} else {
			if err := os.Chmod(config.RuntimeDirectoryPath, 0700); err != nil {
				logger.Fatal(err, "Failed to chmod runtime directory")
			}
			owner := fmt.Sprintf("%d:%d", uid, gid)
			cmd := exec.Command("chown", "-R", owner, config.RuntimeDirectoryPath)
			if err := cmd.Run(); err != nil {
				logger.Fatal(err, "Failed to chown runtime directory")
			}
		}
	}
	owner := fmt.Sprintf("%d:%d", uid, gid)
	cmd := exec.Command("chown", "-R", owner, config.ContainerCwdPath)
	if err := cmd.Run(); err != nil {
		logger.Fatal(err, "Failed to chown container cwd")
	}
	if err := syscall.Setgid(gid); err != nil {
		logger.Fatal(err, "Failed to set GID")
	}
	if err := syscall.Setuid(uid); err != nil {
		logger.Fatal(err, "Failed to set UID")
	}
}
