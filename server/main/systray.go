//go:build windows || linux

package main

import (
	"context"
	"runtime"

	"fyne.io/systray"
)

// TODO: Add macOS support

func runSystray(webUiAddress string, iconBytes []byte, shutdown context.CancelFunc) {
	runtime.LockOSThread()
	systray.Run(func() {
		systray.SetIcon(iconBytes)
		systray.SetTitle("DRPP")
		systray.SetTooltip("DRPP")
		webUiButton := systray.AddMenuItem("Web UI ("+webUiAddress+")", "Launch the web user interface")
		systray.AddSeparator()
		quitButton := systray.AddMenuItem("Quit", "Quit the application")
		go func() {
			for {
				select {
				case _, open := <-webUiButton.ClickedCh:
					if !open {
						return
					}
					go launchUrl(webUiAddress)
				case _, open := <-quitButton.ClickedCh:
					if !open {
						return
					}
					shutdown()
				}
			}
		}()
	}, nil)
}
