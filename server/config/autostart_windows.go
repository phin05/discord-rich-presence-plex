//go:build windows

package config

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const (
	registryKeyPath       = `Software\Microsoft\Windows\CurrentVersion\Run`
	registryValueName     = "DRPP"
	disableWebUiLaunchArg = "--disable-web-ui-launch"
)

var (
	exe = func() string {
		exe, err := os.Executable()
		if err != nil {
			panic(err)
		}
		return exe
	}()
	args = func() string {
		args := os.Args[1:]
		if !slices.Contains(args, disableWebUiLaunchArg) {
			args = append([]string{disableWebUiLaunchArg}, args...)
		}
		return strings.Join(args, " ")
	}()
)

func isAutostartEnabled() (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, registryKeyPath, registry.QUERY_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("open registry key: %w", err)
	}
	defer key.Close()
	value, _, err := key.GetStringValue(registryValueName)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("get registry value: %w", err)
	}
	argsStartPos := 1 + len(exe) + 2
	if len(value) < argsStartPos {
		return false, nil
	}
	currentExe := value[1:][:len(exe)]
	currentArgs := value[argsStartPos:]
	return strings.EqualFold(currentExe, exe) && currentArgs == args, nil
}

func setAutostartEnabled(enabled bool) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, registryKeyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open registry key: %w", err)
	}
	defer key.Close()
	if enabled {
		return key.SetStringValue(registryValueName, fmt.Sprintf(`"%s" %s`, exe, args))
	}
	if err := key.DeleteValue(registryValueName); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return err
	}
	return nil
}
