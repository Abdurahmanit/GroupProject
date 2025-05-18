package config

import (
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	GRPC  GRPCConfig  `mapstructure:"grpc"`
	Mongo MongoConfig `mapstructure:"mongo"` // ИЗМЕНЕНО
	NATS  NATSConfig  `mapstructure:"nats"`
	Redis RedisConfig `mapstructure:"redis"`
}

type GRPCConfig struct {
	Port           string        `mapstructure:"port"`
	MaxRecvMsgSize int           `mapstructure:"max_recv_msg_size"`
	MaxSendMsgSize int           `mapstructure:"max_send_msg_size"`
	Timeout        time.Duration `mapstructure:"timeout"`
}

// MongoConfig хранит конфигурацию для подключения к MongoDB.
type MongoConfig struct {
	URI      string `mapstructure:"uri"`      // Например, "mongodb://localhost:27017"
	Username string `mapstructure:"username"` // Опционально
	Password string `mapstructure:"password"` // Опционально
	Database string `mapstructure:"database"` // Имя базы данных
	// Опционально можно добавить параметры для пула соединений, таймауты и т.д.
	ConnectTimeout time.Duration `mapstructure:"connect_timeout"`
	MinPoolSize    uint64        `mapstructure:"min_pool_size"`
	MaxPoolSize    uint64        `mapstructure:"max_pool_size"`
}

type NATSConfig struct {
	URL            string        `mapstructure:"url"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout"`
}

type RedisConfig struct {
	Address  string `mapstructure:"address"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetDefault("grpc.port", "50051")
	viper.SetDefault("grpc.max_recv_msg_size", 4194304)
	viper.SetDefault("grpc.max_send_msg_size", 4194304)
	viper.SetDefault("grpc.timeout", "10s")

	// Значения по умолчанию для MongoDB
	viper.SetDefault("mongo.uri", "mongodb://localhost:27017")
	viper.SetDefault("mongo.database", "news_service_db") // Пример имени БД
	viper.SetDefault("mongo.connect_timeout", "10s")
	viper.SetDefault("mongo.min_pool_size", 0)   // 0 - без ограничения или значение по умолчанию драйвера
	viper.SetDefault("mongo.max_pool_size", 100) // Значение по умолчанию драйвера

	viper.SetDefault("nats.url", "nats://localhost:4222")
	viper.SetDefault("nats.connect_timeout", "5s")

	viper.SetDefault("redis.address", "localhost:6379")
	viper.SetDefault("redis.db", 0)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		if fi, _ := os.Stat(path); !fi.IsDir() {
			viper.SetConfigFile(path)
		} else {
			viper.AddConfigPath(path)
			viper.SetConfigName("config")
			viper.SetConfigType("yaml")
		}
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("NEWS") // Например, NEWS_MONGO_URI

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found; using defaults and environment variables.")
		} else {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
