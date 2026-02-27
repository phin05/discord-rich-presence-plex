package plex

import (
	"drpp/server/logger"
	"fmt"
)

type prefixedLogger struct {
	prefix string
}

func newPrefixedLogger(prefix string) *prefixedLogger {
	return &prefixedLogger{prefix: prefix}
}

func (p *prefixedLogger) withPrefix(format string) string {
	return fmt.Sprintf("[%s] %s", p.prefix, format)
}

func (p *prefixedLogger) Debug(format string, args ...any) {
	logger.Debug(p.withPrefix(format), args...)
}

func (p *prefixedLogger) Info(format string, args ...any) {
	logger.Info(p.withPrefix(format), args...)
}

func (p *prefixedLogger) Warning(format string, args ...any) {
	logger.Warning(p.withPrefix(format), args...)
}

func (p *prefixedLogger) Error(err error, format string, args ...any) {
	logger.Error(err, p.withPrefix(format), args...)
}
