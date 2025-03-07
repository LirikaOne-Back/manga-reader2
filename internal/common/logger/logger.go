package logger

import (
	"io"
	"log/slog"
	"os"
	"time"
)

// Logger представляет интерфейс для логирования
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
	WithFields(fields map[string]interface{}) Logger
}

// slogLogger реализует интерфейс Logger с использованием log/slog
type slogLogger struct {
	logger *slog.Logger
}

// NewLogger создает новый экземпляр логгера с указанным уровнем логирования
func NewLogger(level string) Logger {
	var logLevel slog.Level

	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handlerOptions := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(time.Now().Format(time.RFC3339))
			}
			return a
		},
	}

	handler := slog.NewJSONHandler(os.Stdout, handlerOptions)
	logger := slog.New(handler)

	return &slogLogger{
		logger: logger,
	}
}

// NewLoggerWithOutput создает новый экземпляр логгера с указанным writer
func NewLoggerWithOutput(level string, w io.Writer) Logger {
	var logLevel slog.Level

	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handlerOptions := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
	}

	handler := slog.NewJSONHandler(w, handlerOptions)
	logger := slog.New(handler)

	return &slogLogger{
		logger: logger,
	}
}

// Debug логирует сообщение с уровнем Debug
func (l *slogLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

// Info логирует сообщение с уровнем Info
func (l *slogLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Warn логирует сообщение с уровнем Warn
func (l *slogLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

// Error логирует сообщение с уровнем Error
func (l *slogLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// With добавляет дополнительные поля к логгеру
func (l *slogLogger) With(args ...any) Logger {
	return &slogLogger{
		logger: l.logger.With(args...),
	}
}

// WithFields добавляет дополнительные поля к логгеру
func (l *slogLogger) WithFields(fields map[string]interface{}) Logger {
	attrs := make([]any, 0, len(fields)*2)

	for k, v := range fields {
		attrs = append(attrs, k, v)
	}

	return &slogLogger{
		logger: l.logger.With(attrs...),
	}
}
