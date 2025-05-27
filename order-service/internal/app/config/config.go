package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string           `yaml:"env" env:"ENV" env-default:"local"`
	GRPCServer GRPCServerConfig `yaml:"grpc_server"`
	MongoDB    MongoDBConfig    `yaml:"mongo"`
	Redis      RedisConfig      `yaml:"redis"`
	NATS       NATSConfig       `yaml:"nats"`
	Logger     LoggerConfig     `yaml:"logger"`
}

type GRPCServerConfig struct {
	Port              string        `yaml:"port" env:"GRPC_PORT_ORDER_SERVICE" env-default:"50054"`
	Timeout           time.Duration `yaml:"timeout" env-default:"5s"`
	MaxConnectionIdle time.Duration `yaml:"max_connection_idle" env-default:"15m"`
	TimeoutGraceful   time.Duration `yaml:"timeout_graceful_shutdown" env-default:"15s"`
}

type MongoDBConfig struct {
	URI      string `yaml:"uri" env:"MONGO_URI" env-default:"mongodb://localhost:27017"`
	User     string `yaml:"user" env:"MONGO_USER"`
	Password string `yaml:"password" env:"MONGO_PASSWORD"`
	Database string `yaml:"database" env:"MONGO_DATABASE" env-default:"order_service_db"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr" env:"REDIS_ADDR" env-default:"localhost:6379"`
	Password string `yaml:"password" env:"REDIS_PASSWORD"`
	DB       int    `yaml:"db" env:"REDIS_DB" env-default:"0"`
}

type NATSConfig struct {
	URL string `yaml:"url" env:"NATS_URL" env-default:"nats://localhost:4222"`
}

type LoggerConfig struct {
	Level      string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
	Encoding   string `yaml:"encoding" env:"LOG_ENCODING" env-default:"json"`
	TimeFormat string `yaml:"time_format" env:"LOG_TIME_FORMAT" env-default:"2006-01-02T15:04:05.000Z07:00"`
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config
	if path == "" {
		err := cleanenv.ReadEnv(&cfg)
		if err != nil {
			return nil, err
		}
		return &cfg, nil
	}

	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		if _, ok := err.(*os.PathError); ok && path != "" {
			log.Printf("Warning: Config file not found at %s, attempting to load from environment variables only.", path)
			errEnv := cleanenv.ReadEnv(&cfg)
			if errEnv != nil {
				return nil, errEnv
			}
			return &cfg, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH_ORDER_SERVICE")
	if configPath == "" {
		configPath = "config.yaml"
	}
	cfg, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}
	return cfg
}
