package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Port            int    `mapstructure:"PORT"`
	UserServiceHost string `mapstructure:"USER_SERVICE_HOST"`
	UserServicePort int    `mapstructure:"USER_SERVICE_PORT"`
	ListingServiceHost string `mapstructure:"LISTING_SERVICE_HOST"` // Добавлено
	ListingServicePort int    `mapstructure:"LISTING_SERVICE_PORT"` // Добавлено
	JWTSecret       string `mapstructure:"JWT_SECRET"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("PORT", 8080)
	viper.SetDefault("USER_SERVICE_HOST", "localhost")
	viper.SetDefault("USER_SERVICE_PORT", 50051)
	viper.SetDefault("LISTING_SERVICE_HOST", "localhost") // Добавлено
	viper.SetDefault("LISTING_SERVICE_PORT", 50052)      // Добавлено
	viper.SetDefault("JWT_SECRET", "your-secret-key")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}