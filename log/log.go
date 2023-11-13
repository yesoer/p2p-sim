package log

import (
	"fmt"
	"runtime"
	"strconv"
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
func log(colorCode, level string, logLevel LogLevel, optionalErr error, message ...any) {
	logPrefix := "%s[%s]%s "
	if logLevel == ErrorLevel {
		fmt.Printf(logPrefix+"%+v\n", colorCode, level, resetColor, optionalErr)
		fmt.Println(message...)
		return
	}

	fmt.Printf(logPrefix, colorCode, level, resetColor)
	fmt.Println(message...)
}

// Info logs information messages, so anything that may be interesting to the
// end user : application health, system resources etc.
func Info(message ...any) {
	log(greenColor, "INFO", InfoLevel, nil, message...)
}

// Error logs error messages
func Error(err error, message ...any) {
	log(redColor, "ERROR", ErrorLevel, err, message...)
}

// Debug logs debug messages, so anything that gives information about specific
// variables and data flow within the application
func Debug(message ...any) {
	file, line := trace(3)
	trace := "\nCalled from  : " + file + ":" + strconv.Itoa(line)
	message = append(message, trace)
	log(blueColor, "DEBUG", DebugLevel, nil, message...)
}

func trace(depth int) (string, int) {
	pc := make([]uintptr, 5) // Adjust the size as needed
	n := runtime.Callers(0, pc)
	frames := runtime.CallersFrames(pc[:n])

	// Skip the first frames, which is the log file
	f, more := frames.Next()
	for i := 0; more && i < depth; i++ {
		f, more = frames.Next()
		if !more {
			return "", 0
		}
	}

	return f.File, f.Line
}
