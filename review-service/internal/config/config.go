package config

import (
	"os"
	"strconv"

	"github.com/Abdurahmanit/GroupProject/review-service/internal/platform/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap" // For logging within config loading if needed
)

// Config holds all configuration for the service.
type Config struct {
	ServiceName            string `mapstructure:"SERVICE_NAME"`
	GRPCPort               string `mapstructure:"GRPC_PORT"`
	MongoURI               string `mapstructure:"MONGO_URI"`
	MongoDatabase          string `mapstructure:"MONGO_DATABASE"`
	NATSURL                string `mapstructure:"NATS_URL"`
	JWTSecret              string `mapstructure:"JWT_SECRET"` // For validating tokens from API Gateway
	PrometheusMetricsPort  string `mapstructure:"PROMETHEUS_METRICS_PORT"`
	LogLevel               string `mapstructure:"LOG_LEVEL"`
	LogFormat              string `mapstructure:"LOG_FORMAT"`
	OTExporterOTLPEndpoint string `mapstructure:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	// Add other service-specific configurations here
}

// LoadConfig reads configuration from environment variables and/or .env file.
func LoadConfig(appLogger *logger.Logger) (*Config, error) {
	// Set default values
	viper.SetDefault("SERVICE_NAME", "review-service")
	viper.SetDefault("GRPC_PORT", "50053") // Default gRPC port for review-service
	viper.SetDefault("MONGO_URI", "mongodb://localhost:27017")
	viper.SetDefault("MONGO_DATABASE", "bicycle_shop_reviews") // Specific DB for reviews
	viper.SetDefault("NATS_URL", "nats://localhost:4222")
	viper.SetDefault("JWT_SECRET", "your-very-secret-key-for-review-service") // CHANGE THIS!
	viper.SetDefault("PROMETHEUS_METRICS_PORT", "9093")                       // Default metrics port
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORMAT", "json")
	viper.SetDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "") // e.g., "otel-collector:4317"

	// Tell viper to look for environment variables
	viper.AutomaticEnv()
	// Optional: Read from a .env file if present (godotenv is called in main.go)
	// viper.SetConfigName(".env")
	// viper.SetConfigType("env")
	// viper.AddConfigPath(".")
	// if err := viper.ReadInConfig(); err != nil {
	// 	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
	// 		appLogger.Info("No .env config file found, relying on OS environment variables.")
	// 	} else {
	// 		appLogger.Warn("Error reading .env config file", zap.Error(err))
	// 	}
	// }

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		appLogger.Error("Failed to unmarshal configuration", zap.Error(err))
		return nil, err
	}

	// Manual override for specific types if viper struggles with direct env var parsing for all cases
	// For example, if GRPC_PORT was an int, you might do:
	// portStr := getEnv("GRPC_PORT", "50053")
	// cfg.GRPCPort, _ = strconv.Atoi(portStr) // Add error handling

	// Validate critical configurations
	if cfg.JWTSecret == "your-very-secret-key-for-review-service" || cfg.JWTSecret == "" {
		appLogger.Warn("JWT_SECRET is set to its default insecure value or is empty. Please set a strong secret in your environment.")
		// Depending on policy, you might choose to Fatal here if JWTSecret is absolutely mandatory.
	}
	if cfg.MongoURI == "" {
		appLogger.Fatal("MONGO_URI is not set. This is required.")
	}
	if cfg.MongoDatabase == "" {
		appLogger.Fatal("MONGO_DATABASE is not set. This is required.")
	}

	appLogger.Debug("Configuration loaded",
		zap.String("service_name", cfg.ServiceName),
		zap.String("grpc_port", cfg.GRPCPort),
		zap.Bool("mongo_uri_present", cfg.MongoURI != ""),
		zap.String("mongo_database", cfg.MongoDatabase),
		zap.String("nats_url", cfg.NATSURL),
		zap.Bool("jwt_secret_present", cfg.JWTSecret != ""),
		zap.String("prometheus_port", cfg.PrometheusMetricsPort),
		zap.String("log_level", cfg.LogLevel),
		zap.String("log_format", cfg.LogFormat),
		zap.String("otel_endpoint", cfg.OTExporterOTLPEndpoint),
	)

	return &cfg, nil
}

// Helper function to get environment variable with a fallback (already in your logger's config, but can be local too)
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// Helper function to get environment variable as int with a fallback
func getEnvAsInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return fallback
}
