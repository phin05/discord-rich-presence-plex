//go:build !windows

// TODO: Add Linux and macOS support

package config

import "drpp/server/api"

var errUnavailable = api.ErrServiceUnavailable("Autostart is not available for this platform")

func isAutostartEnabled() (bool, error) {
	return false, errUnavailable
}

func setAutostartEnabled(enabled bool) error {
	return errUnavailable
}
