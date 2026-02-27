//go:build !windows

package discord

import (
	"net"
	"time"
)

func dial(path string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("unix", path, timeout)
}
