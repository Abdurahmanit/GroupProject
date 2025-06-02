package config

import (
	"errors"
	"fmt"

	"github.com/Abdurahmanit/GroupProject/review-service/internal/platform/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	ServiceName            string `mapstructure:"SERVICE_NAME"`
	GRPCPort               string `mapstructure:"GRPC_PORT"`
	MongoURI               string `mapstructure:"MONGO_URI"`
	MongoDatabase          string `mapstructure:"MONGO_DATABASE"`
	NATSURL                string `mapstructure:"NATS_URL"`
	JWTSecret              string `mapstructure:"JWT_SECRET"`
	PrometheusMetricsPort  string `mapstructure:"PROMETHEUS_METRICS_PORT"`
	LogLevel               string `mapstructure:"LOG_LEVEL"`
	LogFormat              string `mapstructure:"LOG_FORMAT"`
	OTExporterOTLPEndpoint string `mapstructure:"OTEL_EXPORTER_OTLP_ENDPOINT"`
}

func LoadConfig(appLogger *logger.Logger) (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			appLogger.Info(".env file not found. Relying entirely on environment variables.")
		} else {
			appLogger.Warn("Error reading .env file", zap.Error(err))
		}
	} else {
		appLogger.Info("Successfully loaded configuration from .env file.")
	}

	viper.BindEnv("SERVICE_NAME")
	viper.BindEnv("GRPC_PORT")
	viper.BindEnv("MONGO_URI")
	viper.BindEnv("MONGO_DATABASE")
	viper.BindEnv("NATS_URL")
	viper.BindEnv("JWT_SECRET")
	viper.BindEnv("PROMETHEUS_METRICS_PORT")
	viper.BindEnv("LOG_LEVEL")
	viper.BindEnv("LOG_FORMAT")
	viper.BindEnv("OTEL_EXPORTER_OTLP_ENDPOINT")

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		appLogger.Error("Failed to unmarshal configuration", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	if cfg.GRPCPort == "" {
		errMsg := "critical configuration GRPC_PORT is not set"
		appLogger.Error(errMsg)
		return nil, errors.New(errMsg)
	}
	if cfg.MongoURI == "" {
		errMsg := "critical configuration MONGO_URI is not set"
		appLogger.Error(errMsg)
		return nil, errors.New(errMsg)
	}
	if cfg.MongoDatabase == "" {
		errMsg := "critical configuration MONGO_DATABASE is not set"
		appLogger.Error(errMsg)
		return nil, errors.New(errMsg)
	}
	if cfg.JWTSecret == "" {
		errMsg := "critical security configuration JWT_SECRET is not set"
		appLogger.Error(errMsg)
		return nil, errors.New(errMsg)
	}

	if cfg.ServiceName == "" {
		appLogger.Warn("SERVICE_NAME is not set in .env or environment variables. Defaulting to 'review-service'.")
		cfg.ServiceName = "review-service"
	}
	if cfg.LogLevel == "" {
		appLogger.Warn("LOG_LEVEL is not set in .env or environment variables. Defaulting to 'info'.")
		cfg.LogLevel = "info"
	}
	if cfg.LogFormat == "" {
		appLogger.Warn("LOG_FORMAT is not set in .env or environment variables. Defaulting to 'json'.")
		cfg.LogFormat = "json"
	}
	if cfg.NATSURL == "" {
		appLogger.Warn("NATS_URL is not set. NATS-dependent features may be unavailable or the application may fail if NATS is required.")
	}
	if cfg.PrometheusMetricsPort == "" {
		appLogger.Info("PROMETHEUS_METRICS_PORT is not set. Prometheus metrics server will not start.")
	}

	appLogger.Debug("Configuration loaded successfully",
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
