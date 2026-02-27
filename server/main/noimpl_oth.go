//go:build !(windows || linux)

package main

import "context"

func setupContainer() {}

func runSystray(webUiAddress string, iconBytes []byte, shutdown context.CancelFunc) {}
