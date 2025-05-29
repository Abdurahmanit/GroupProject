package config

import (
	"log" // Using log for simplicity in config loading status/errors

	"github.com/spf13/viper"
)

type Config struct {
	Port               int    `mapstructure:"PORT"`
	UserServiceHost    string `mapstructure:"USER_SERVICE_HOST"`
	UserServicePort    int    `mapstructure:"USER_SERVICE_PORT"`
	ListingServiceHost string `mapstructure:"LISTING_SERVICE_HOST"`
	ListingServicePort int    `mapstructure:"LISTING_SERVICE_PORT"`
	JWTSecret          string `mapstructure:"JWT_SECRET"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Info: Error reading .env config file: %s. Will rely on OS environment variables if set.\n", err)
	}

	viper.BindEnv("PORT", "PORT")
	viper.BindEnv("USER_SERVICE_HOST", "USER_SERVICE_HOST")
	viper.BindEnv("USER_SERVICE_PORT", "USER_SERVICE_PORT")
	viper.BindEnv("LISTING_SERVICE_HOST", "LISTING_SERVICE_HOST")
	viper.BindEnv("LISTING_SERVICE_PORT", "LISTING_SERVICE_PORT")
	viper.BindEnv("JWT_SECRET", "JWT_SECRET")
	viper.AutomaticEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	log.Printf("API Gateway configuration loaded. PORT resolved to: %d\n", cfg.Port)

	if cfg.Port == 0 {
		log.Println("Warning: API Gateway PORT is 0 after loading configuration. Please check your .env file and environment variable settings for 'PORT'.")
	}

	return &cfg, nil
}
