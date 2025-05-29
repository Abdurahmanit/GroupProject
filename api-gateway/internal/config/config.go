package config

import (
	"github.com/spf13/viper"
)

// Config структура для хранения конфигурации API Gateway
type Config struct {
	Port               int    `mapstructure:"PORT"`
	UserServiceHost    string `mapstructure:"USER_SERVICE_HOST"`
	UserServicePort    int    `mapstructure:"USER_SERVICE_PORT"`
	ListingServiceHost string `mapstructure:"LISTING_SERVICE_HOST"`
	ListingServicePort int    `mapstructure:"LISTING_SERVICE_PORT"`
	JWTSecret          string `mapstructure:"JWT_SECRET"`
	// Добавь сюда хосты и порты для других сервисов по мере их подключения
}

// LoadConfig загружает конфигурацию из .env файла и переменных окружения
func LoadConfig() (*Config, error) {
	// Явное указание Viper'у на переменные окружения
	viper.BindEnv("port", "PORT")
	viper.BindEnv("user_service_host", "USER_SERVICE_HOST")
	viper.BindEnv("user_service_port", "USER_SERVICE_PORT")
	viper.BindEnv("listing_service_host", "LISTING_SERVICE_HOST")
	viper.BindEnv("listing_service_port", "LISTING_SERVICE_PORT")
	viper.BindEnv("jwt_secret", "JWT_SECRET")
	// Добавь сюда BindEnv для других сервисов

	viper.AutomaticEnv() // Позволяет Viper также подхватывать другие переменные окружения

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
