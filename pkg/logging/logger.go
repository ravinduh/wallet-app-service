package logging

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ravindu/wallet-app-service/pkg/request"
)

// Logger represents a logger instance
type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
	warnLogger  *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	return &Logger{
		infoLogger:  log.New(os.Stdout, "INFO: ", log.LstdFlags),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.LstdFlags|log.Lshortfile),
		debugLogger: log.New(os.Stdout, "DEBUG: ", log.LstdFlags|log.Lshortfile),
		warnLogger:  log.New(os.Stdout, "WARN: ", log.LstdFlags),
	}
}

// getRequestID extracts request ID from context
func getRequestID(ctx context.Context) string {
	if ctx == nil {
		return "no-request-id"
	}

	requestID, ok := ctx.Value(request.RequestIDKey).(string)
	if !ok || requestID == "" {
		return "no-request-id"
	}

	return requestID
}

// formatMessage formats the log message with timestamp and request ID
func formatMessage(ctx context.Context, message string) string {
	requestID := getRequestID(ctx)
	timestamp := time.Now().Format(time.RFC3339)
	return timestamp + " [" + requestID + "] " + message
}

// Info logs informational messages
func (l *Logger) Info(ctx context.Context, message string) {
	l.infoLogger.Println(formatMessage(ctx, message))
}

// Error logs error messages
func (l *Logger) Error(ctx context.Context, message string) {
	l.errorLogger.Println(formatMessage(ctx, message))
}

// Warn logs warning messages
func (l *Logger) Warn(ctx context.Context, message string) {
	l.warnLogger.Println(formatMessage(ctx, message))
}

// Debug logs debug messages
func (l *Logger) Debug(ctx context.Context, message string) {
	l.debugLogger.Println(formatMessage(ctx, message))
}

// With returns a Logger with the given key value pair for structured logging
func (l *Logger) With(ctx context.Context, key string, value interface{}) *Logger {
	// Currently a simple implementation; could be extended for more complex structured logging
	l.Debug(ctx, key+": "+toString(value))
	return l
}

// toString converts a value to a string
func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int64, uint, uint64, float64:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%+v", v)
	}
}