package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Port      int    `mapstructure:"PORT"`
	MongoURI  string `mapstructure:"MONGO_URI"`
	RedisAddr string `mapstructure:"REDIS_ADDR"`
	JWTSecret string `mapstructure:"JWT_SECRET"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("PORT", 50051)
	viper.SetDefault("MONGO_URI", "mongodb://localhost:27017")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
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
