//go:build windows

package config

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

const (
	registryKeyPath   = `Software\Microsoft\Windows\CurrentVersion\Run`
	registryValueName = "DRPP"
)

var command = func() string {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf(`"%s" --disable-web-ui-launch`, exe)
}()

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
	if errors.Is(err, registry.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("get registry value: %w", err)
	}
	return value == command, nil
}

func setAutostartEnabled(enabled bool) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, registryKeyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open registry key: %w", err)
	}
	defer key.Close()
	if enabled {
		return key.SetStringValue(registryValueName, command)
	}
	if err := key.DeleteValue(registryValueName); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return err
	}
	return nil
}
