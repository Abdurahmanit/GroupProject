package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
	config *LoggerConfig
}

var (
	globalLogger *Logger
	once         sync.Once
)

func NewLogger() *Logger {
	once.Do(func() {
		cfg := DefaultConfig()

		var zapConfig zap.Config
		if cfg.Level == "debug" {
			zapConfig = zap.NewDevelopmentConfig()
			zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Colored level for development
		} else { // Production-ready config
			zapConfig = zap.NewProductionConfig()
			zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // Standard time format
		}

		// Set the log level
		err := zapConfig.Level.UnmarshalText([]byte(cfg.Level))
		if err != nil {
			// Fallback to info level if parsing fails
			fmt.Fprintf(os.Stderr, "Warning: Invalid LOG_LEVEL '%s', defaulting to 'info'. Error: %v\n", cfg.Level, err)
			zapConfig.Level.SetLevel(zapcore.InfoLevel)
		}

		// Configure output paths
		if cfg.OutputFile != "stdout" && cfg.OutputFile != "stderr" {
			// Ensure the directory for the log file exists
			logDir := filepath.Dir(cfg.OutputFile)
			if err := os.MkdirAll(logDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to create log directory '%s', defaulting to stdout. Error: %v\n", logDir, err)
				zapConfig.OutputPaths = []string{"stdout"}
				zapConfig.ErrorOutputPaths = []string{"stderr"}
			} else {
				zapConfig.OutputPaths = []string{cfg.OutputFile, "stdout"} // Log to file and stdout
				zapConfig.ErrorOutputPaths = []string{cfg.OutputFile, "stderr"}
			}
		} else {
			zapConfig.OutputPaths = []string{cfg.OutputFile}
			zapConfig.ErrorOutputPaths = []string{"stderr"}
		}

		// Set encoder based on format
		if strings.ToLower(cfg.Format) == "console" || strings.ToLower(cfg.Format) == "text" {
			zapConfig.Encoding = "console"
			zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Example: colored level for console
		} else { // Default to JSON
			zapConfig.Encoding = "json"
		}

		logger, err := zapConfig.Build(zap.AddCallerSkip(1)) // AddCallerSkip to show correct caller
		if err != nil {
			// Fallback to a basic Zap logger if custom configuration fails
			fmt.Fprintf(os.Stderr, "Error initializing custom Zap logger: %v. Falling back to basic logger.\n", err)
			logger, _ = zap.NewProduction()
		}

		globalLogger = &Logger{Logger: logger, config: cfg} // Store the configured logger
		globalLogger.Info("Logger initialized", zap.String("level", cfg.Level), zap.String("format", cfg.Format), zap.Strings("output_paths", zapConfig.OutputPaths))
	})
	return globalLogger
}

func (l *Logger) Named(name string) *Logger {
	return &Logger{Logger: l.Logger.Named(name), config: l.config}
}

func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{Logger: l.Logger.With(fields...), config: l.config}
}
