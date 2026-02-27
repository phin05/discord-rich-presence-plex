//go:build windows

package discord

import (
	"net"
	"time"

	"github.com/Microsoft/go-winio"
)

func dial(path string, timeout time.Duration) (net.Conn, error) {
	return winio.DialPipe(path, &timeout)
}
