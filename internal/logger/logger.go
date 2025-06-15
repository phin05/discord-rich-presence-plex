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

func log(w io.Writer, prefixColour string, prefix string, format string, args ...any) {
	_, source, _, ok := runtime.Caller(2)
	if ok {
		sourceSplit := strings.Split(source, "/")
		if sourceSplit[len(sourceSplit)-1] == "main.go" {
			source = "main"
		} else {
			source = sourceSplit[len(sourceSplit)-2]
		}
	} else {
		source = "?"
	}
	fmt.Fprintln(w, fmt.Sprintf(colourGray+time.Now().Format(time.RFC3339)+colourReset+" ["+prefixColour+prefix+colourReset+"] ["+colourBlue+source+colourReset+"] "+format, args...))
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

func Error(err error, format string, args ...any) {
	log(os.Stderr, colourRed, "ERROR", fmt.Sprintf("%s: %v", format, err), args...)
}

func Fatal(err error, format string, args ...any) {
	log(os.Stderr, colourRed, "FATAL", fmt.Sprintf("%s: %v", format, err), args...)
	os.Exit(1)
}
