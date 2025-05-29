package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Port                int    `mapstructure:"PORT"`
	MongoURI            string `mapstructure:"MONGO_URI"`
	RedisAddr           string `mapstructure:"REDIS_ADDR"`
	JWTSecret           string `mapstructure:"JWT_SECRET"`
	MailerSendAPIKey    string `mapstructure:"MAILERSEND_API_KEY"`
	MailerSendFromEmail string `mapstructure:"MAILERSEND_FROM_EMAIL"`
	MailerSendFromName  string `mapstructure:"MAILERSEND_FROM_NAME"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")        // For local development, e.g., .env file
	viper.AddConfigPath("./config") // For Dockerized environments
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("PORT", 50051)
	viper.SetDefault("MONGO_URI", "mongodb://localhost:27017")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("JWT_SECRET", "your-secret-key")
	viper.SetDefault("MAILERSEND_API_KEY", "your-mailersend-api-key")
	viper.SetDefault("MAILERSEND_FROM_EMAIL", "noreply@example.com")
	viper.SetDefault("MAILERSEND_FROM_NAME", "Your Application Name")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file not found error is ignored if not present, env vars will be used.
			// Other errors should be returned.
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
