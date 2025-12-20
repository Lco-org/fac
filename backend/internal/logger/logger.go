// Package logger provides structured logging for CSO-ERP
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

// Level represents log levels
type Level int

const (
	// LevelDebug for debug messages
	LevelDebug Level = iota
	// LevelInfo for informational messages
	LevelInfo
	// LevelWarn for warning messages
	LevelWarn
	// LevelError for error messages
	LevelError
	// LevelFatal for fatal messages
	LevelFatal
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp  string                 `json:"timestamp"`
	Level      string                 `json:"level"`
	Message    string                 `json:"message"`
	RequestID  string                 `json:"request_id,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	TenantDB   string                 `json:"tenant_db,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	StatusCode int                    `json:"status_code,omitempty"`
	Duration   string                 `json:"duration,omitempty"`
	IP         string                 `json:"ip,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	Error      string                 `json:"error,omitempty"`
	ErrorCode  string                 `json:"error_code,omitempty"`
	Stack      string                 `json:"stack,omitempty"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	Caller     string                 `json:"caller,omitempty"`
}

// Logger is the main logger struct
type Logger struct {
	mu        sync.Mutex
	output    io.Writer
	level     Level
	requestID string
	userID    string
	tenantDB  string
	fields    map[string]interface{}
}

// Config holds logger configuration
type Config struct {
	Level       Level
	Output      io.Writer
	JSONFormat  bool
	EnableColor bool
}

// DefaultConfig returns default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:       LevelInfo,
		Output:      os.Stdout,
		JSONFormat:  true,
		EnableColor: false,
	}
}

// Global logger instance
var (
	defaultLogger *Logger
	once          sync.Once
)

// Init initializes the global logger
func Init(cfg Config) {
	once.Do(func() {
		defaultLogger = New(cfg)
	})
}

// New creates a new logger instance
func New(cfg Config) *Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}
	return &Logger{
		output: cfg.Output,
		level:  cfg.Level,
		fields: make(map[string]interface{}),
	}
}

// Default returns the default logger
func Default() *Logger {
	if defaultLogger == nil {
		Init(DefaultConfig())
	}
	return defaultLogger
}

// WithRequestID adds request ID to logger
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		output:    l.output,
		level:     l.level,
		requestID: requestID,
		userID:    l.userID,
		tenantDB:  l.tenantDB,
		fields:    copyFields(l.fields),
	}
}

// WithUserID adds user ID to logger
func (l *Logger) WithUserID(userID string) *Logger {
	return &Logger{
		output:    l.output,
		level:     l.level,
		requestID: l.requestID,
		userID:    userID,
		tenantDB:  l.tenantDB,
		fields:    copyFields(l.fields),
	}
}

// WithTenant adds tenant database to logger
func (l *Logger) WithTenant(tenantDB string) *Logger {
	return &Logger{
		output:    l.output,
		level:     l.level,
		requestID: l.requestID,
		userID:    l.userID,
		tenantDB:  tenantDB,
		fields:    copyFields(l.fields),
	}
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newFields := copyFields(l.fields)
	newFields[key] = value
	return &Logger{
		output:    l.output,
		level:     l.level,
		requestID: l.requestID,
		userID:    l.userID,
		tenantDB:  l.tenantDB,
		fields:    newFields,
	}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newFields := copyFields(l.fields)
	for k, v := range fields {
		newFields[k] = v
	}
	return &Logger{
		output:    l.output,
		level:     l.level,
		requestID: l.requestID,
		userID:    l.userID,
		tenantDB:  l.tenantDB,
		fields:    newFields,
	}
}

// WithError adds error to the logger
func (l *Logger) WithError(err error) *Logger {
	if err == nil {
		return l
	}
	return l.WithField("error", err.Error())
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.log(LevelDebug, msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.log(LevelInfo, msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.log(LevelWarn, msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	l.log(LevelError, msg, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.log(LevelFatal, msg, args...)
	os.Exit(1)
}

func (l *Logger) log(level Level, msg string, args ...interface{}) {
	if level < l.level {
		return
	}

	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level.String(),
		Message:   msg,
		RequestID: l.requestID,
		UserID:    l.userID,
		TenantDB:  l.tenantDB,
		Fields:    l.fields,
		Caller:    getCaller(3),
	}

	if errVal, ok := l.fields["error"]; ok {
		entry.Error = fmt.Sprint(errVal)
		delete(entry.Fields, "error")
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(l.output, "failed to marshal log entry: %v\n", err)
		return
	}

	fmt.Fprintln(l.output, string(data))
}

// LogRequest logs an HTTP request
func (l *Logger) LogRequest(method, path string, statusCode int, duration time.Duration, ip, userAgent string) {
	entry := LogEntry{
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Level:      LevelInfo.String(),
		Message:    "HTTP Request",
		RequestID:  l.requestID,
		UserID:     l.userID,
		TenantDB:   l.tenantDB,
		Method:     method,
		Path:       path,
		StatusCode: statusCode,
		Duration:   duration.String(),
		IP:         ip,
		UserAgent:  userAgent,
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	data, _ := json.Marshal(entry)
	fmt.Fprintln(l.output, string(data))
}

// LogError logs an error with stack trace
func (l *Logger) LogError(err error, code string, msg string) {
	if err == nil {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     LevelError.String(),
		Message:   msg,
		RequestID: l.requestID,
		UserID:    l.userID,
		TenantDB:  l.tenantDB,
		Error:     err.Error(),
		ErrorCode: code,
		Stack:     getStackTrace(),
		Caller:    getCaller(2),
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	data, _ := json.Marshal(entry)
	fmt.Fprintln(l.output, string(data))
}

// Helper functions

func copyFields(fields map[string]interface{}) map[string]interface{} {
	newFields := make(map[string]interface{})
	for k, v := range fields {
		newFields[k] = v
	}
	return newFields
}

func getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	// Get only the filename
	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' || file[i] == '\\' {
			short = file[i+1:]
			break
		}
	}
	return fmt.Sprintf("%s:%d", short, line)
}

func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// ============================================================================
// GLOBAL CONVENIENCE FUNCTIONS
// ============================================================================

// Debug logs a debug message using the default logger
func Debug(msg string, args ...interface{}) {
	Default().Debug(msg, args...)
}

// Info logs an info message using the default logger
func Info(msg string, args ...interface{}) {
	Default().Info(msg, args...)
}

// Warn logs a warning message using the default logger
func Warn(msg string, args ...interface{}) {
	Default().Warn(msg, args...)
}

// Error logs an error message using the default logger
func Error(msg string, args ...interface{}) {
	Default().Error(msg, args...)
}

// Fatal logs a fatal message using the default logger
func Fatal(msg string, args ...interface{}) {
	Default().Fatal(msg, args...)
}

// WithRequestID creates a new logger with request ID
func WithRequestID(requestID string) *Logger {
	return Default().WithRequestID(requestID)
}

// WithUserID creates a new logger with user ID
func WithUserID(userID string) *Logger {
	return Default().WithUserID(userID)
}

// WithField creates a new logger with a field
func WithField(key string, value interface{}) *Logger {
	return Default().WithField(key, value)
}

// WithError creates a new logger with an error
func WithError(err error) *Logger {
	return Default().WithError(err)
}
