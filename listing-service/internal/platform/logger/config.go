package logger

import (
	"os"
	"strings"
)

type LoggerConfig struct {
	Level  string // "info", "error", "debug"
	Format string // "json", "text"
}

func DefaultConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:  getEnv("LOG_LEVEL", "info"),
		Format: getEnv("LOG_FORMAT", "json"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func (c *LoggerConfig) ShouldLog(level string) bool {
	levels := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2, // Добавлено для поддержки warn
		"error": 3,
	}
	return levels[strings.ToLower(level)] >= levels[strings.ToLower(c.Level)]
}