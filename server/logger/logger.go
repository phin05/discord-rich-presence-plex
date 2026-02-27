package logger

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	colourGray    = "\033[37m"
	colourMagenta = "\033[35m"
	colourGreen   = "\033[32m"
	colourYellow  = "\033[33m"
	colourRed     = "\033[31m"
	colourBlue    = "\033[34m"
	colourReset   = "\033[0m"
)

var EnableDebugOutput atomic.Bool

var (
	logFile   *os.File
	logFileMu sync.Mutex
)

func SetLogFile(path string) error {
	logFileMu.Lock()
	defer logFileMu.Unlock()
	if logFile != nil {
		return errors.New("log file already set")
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600) //nolint:gosec
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	if err := debug.SetCrashOutput(file, debug.CrashOptions{}); err != nil {
		return fmt.Errorf("set crash output: %w", err)
	}
	logFile = file
	return nil
}

func CloseLogFile() error {
	logFileMu.Lock()
	defer logFileMu.Unlock()
	if logFile == nil {
		return errors.New("log file not set")
	}
	return logFile.Close()
}

var entryId atomic.Uint64

func log(levelColour string, level string, err error, format string, args ...any) {
	_, source, _, ok := runtime.Caller(2)
	if ok {
		// Source is guaranteed to have forward slashes (even on Windows) and at least two segments (package and file)
		sourceSplit := strings.Split(source, "/")
		source = sourceSplit[len(sourceSplit)-2]
	} else {
		source = "unknown"
	}
	timestamp := time.Now().Format(time.RFC3339)
	message := fmt.Sprintf(format, args...)
	if err != nil {
		message += ": " + err.Error()
	}
	_, _ = fmt.Fprintf(os.Stderr, "%s%s%s [%s%s%s] [%s%s%s] %s\n", colourGray, timestamp, colourReset, levelColour, level, colourReset, colourBlue, source, colourReset, message)
	writeToLogFile(timestamp, level, source, message)
	entry := entry{
		Id:        entryId.Add(1),
		Timestamp: timestamp,
		Level:     level,
		Source:    source,
		Message:   message,
	}
	buffer.add(entry)
	notifySubs(entry)
}

func writeToLogFile(timestamp string, level string, source string, message string) {
	logFileMu.Lock()
	defer logFileMu.Unlock()
	if logFile != nil {
		if _, err := fmt.Fprintf(logFile, "%s [%s] [%s] %s\n", timestamp, level, source, message); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to write to log file: %v\n", err)
		}
	}
}

func Debug(format string, args ...any) {
	if EnableDebugOutput.Load() {
		log(colourMagenta, "DEBUG", nil, format, args...)
	}
}

func Info(format string, args ...any) {
	log(colourGreen, "INFO", nil, format, args...)
}

func Warning(format string, args ...any) {
	log(colourYellow, "WARN", nil, format, args...)
}

func Error(err error, format string, args ...any) {
	log(colourRed, "ERROR", err, format, args...)
}

func Fatal(err error, format string, args ...any) {
	log(colourRed, "FATAL", err, format, args...)
	_ = CloseLogFile()
	os.Exit(1)
}
