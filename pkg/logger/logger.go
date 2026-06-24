package logger

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Level represents log level
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel parses a log level string
func ParseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// Logger is a simple structured logger
type Logger struct {
	level   Level
	format  string // "text" or "json"
	output  *os.File
	noColor bool
}

// New creates a new logger
func New(level, format string, noColor bool) *Logger {
	return &Logger{
		level:   ParseLevel(level),
		format:  format,
		output:  os.Stderr,
		noColor: noColor,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...Field) {
	l.log(DebugLevel, msg, fields...)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...Field) {
	l.log(InfoLevel, msg, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...Field) {
	l.log(WarnLevel, msg, fields...)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...Field) {
	l.log(ErrorLevel, msg, fields...)
}

func (l *Logger) log(level Level, msg string, fields ...Field) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02T15:04:05Z07:00")
	
	if l.format == "json" {
		l.logJSON(timestamp, level, msg, fields)
	} else {
		l.logText(timestamp, level, msg, fields)
	}
}

func (l *Logger) logText(timestamp string, level Level, msg string, fields []Field) {
	levelStr := level.String()
	if !l.noColor {
		switch level {
		case DebugLevel:
			levelStr = "\033[36m" + levelStr + "\033[0m"
		case InfoLevel:
			levelStr = "\033[32m" + levelStr + "\033[0m"
		case WarnLevel:
			levelStr = "\033[33m" + levelStr + "\033[0m"
		case ErrorLevel:
			levelStr = "\033[31m" + levelStr + "\033[0m"
		}
	}

	fmt.Fprintf(l.output, "[%s] %s %s", timestamp, levelStr, msg)
	for _, f := range fields {
		fmt.Fprintf(l.output, " %s=%v", f.Key, f.Value)
	}
	fmt.Fprintln(l.output)
}

func (l *Logger) logJSON(timestamp string, level Level, msg string, fields []Field) {
	pairs := []string{
		fmt.Sprintf(`"time":"%s"`, timestamp),
		fmt.Sprintf(`"level":"%s"`, level.String()),
		fmt.Sprintf(`"msg":"%s"`, msg),
	}
	for _, f := range fields {
		pairs = append(pairs, fmt.Sprintf(`"%s":"%v"`, f.Key, f.Value))
	}
	fmt.Fprintf(l.output, "{%s}\n", strings.Join(pairs, ","))
}

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// F creates a new field
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Default logger instance
var Default = New("info", "text", false)
