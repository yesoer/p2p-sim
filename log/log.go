package log

import (
	"fmt"
)

const (
	resetColor = "\033[0m"
	redColor   = "\033[31m"
	greenColor = "\033[32m"
	blueColor  = "\033[34m"
)

// LogLevel represents the log level
type LogLevel int

const (
	// InfoLevel represents the info log level
	InfoLevel LogLevel = iota
	// ErrorLevel represents the error log level
	ErrorLevel
	// DebugLevel represents the debug log level
	DebugLevel
)

// log prints the message with the specified color
func log(colorCode, level, message string, logLevel LogLevel) {
	if logLevel >= InfoLevel {
		fmt.Printf("%s[%s]%s %s\n", colorCode, level, resetColor, message)
	}
}

// Info logs information messages
func Info(message string) {
	log(greenColor, "INFO", message, InfoLevel)
}

// Error logs error messages
func Error(message string) {
	log(redColor, "ERROR", message, ErrorLevel)
}

// Debug logs debug messages
func Debug(message string) {
	log(blueColor, "DEBUG", message, DebugLevel)
}
