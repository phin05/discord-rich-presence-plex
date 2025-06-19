package logger

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
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

func log(w io.Writer, levelColour string, level string, format string, args ...any) {
	_, source, _, ok := runtime.Caller(2)
	if ok {
		sourceSplit := strings.Split(source, "/")
		if sourceSplit[len(sourceSplit)-1] == "main.go" {
			source = "main.go"
		} else {
			source = strings.Join(sourceSplit[len(sourceSplit)-2:], "/")
		}
	} else {
		source = "?"
	}
	fmt.Fprintln(w, fmt.Sprintf(colourGray+time.Now().Format(time.RFC3339)+colourReset+" ["+levelColour+level+colourReset+"] ["+colourBlue+source+colourReset+"] "+format, args...))
}

func Debug(format string, args ...any) {
	log(os.Stderr, colourMagenta, "DEBUG", format, args...)
}

func Info(format string, args ...any) {
	log(os.Stdout, colourGreen, "INFO", format, args...)
}

func Warning(format string, args ...any) {
	log(os.Stderr, colourYellow, "WARN", format, args...)
}

func appendError(err error, format string) string {
	if err == nil {
		return format
	}
	return format + ": " + err.Error()
}

func Error(err error, format string, args ...any) {
	log(os.Stderr, colourRed, "ERROR", appendError(err, format), args...)
}

func Fatal(err error, format string, args ...any) {
	log(os.Stderr, colourRed, "FATAL", appendError(err, format), args...)
	os.Exit(1)
}
