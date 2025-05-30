package logger

import (
	"os"
	"strings"

	"go.uber.org/zap/zapcore"
)

// LoggerConfig holds configuration for the logger.
type LoggerConfig struct {
	Level      string
	Format     string
	OutputFile string
}

// getEnv is a helper to read an environment variable or return a default value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// DefaultConfig creates a new LoggerConfig with default values, typically read from environment variables.
func DefaultConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:      strings.ToLower(getEnv("LOG_LEVEL", "info")),
		Format:     strings.ToLower(getEnv("LOG_FORMAT", "json")),
		OutputFile: getEnv("LOG_OUTPUT_FILE", "stdout"), // Default to stdout
	}
}

// ToZapLevel converts the string log level to zapcore.Level.
func (c *LoggerConfig) ToZapLevel() zapcore.Level {
	switch c.Level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	case "panic":
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel // Default to Info if unspecified or invalid
	}
}
