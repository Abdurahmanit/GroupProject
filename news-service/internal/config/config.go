package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type SMTPConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	SenderEmail string `mapstructure:"sender_email"`
}

type Config struct {
	GRPC               GRPCConfig  `mapstructure:"grpc"`
	Mongo              MongoConfig `mapstructure:"mongo"`
	NATS               NATSConfig  `mapstructure:"nats"`
	Redis              RedisConfig `mapstructure:"redis"`
	SMTP               SMTPConfig  `mapstructure:"smtp"`
	UserServiceAddress string      `mapstructure:"user_service_address"`
}

type GRPCConfig struct {
	Port           string        `mapstructure:"port"`
	MaxRecvMsgSize int           `mapstructure:"max_recv_msg_size"`
	MaxSendMsgSize int           `mapstructure:"max_send_msg_size"`
	Timeout        time.Duration `mapstructure:"timeout"`
}

type MongoConfig struct {
	URI            string        `mapstructure:"uri"`
	Username       string        `mapstructure:"username"`
	Password       string        `mapstructure:"password"`
	Database       string        `mapstructure:"database"`
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
	viper.SetDefault("grpc.port", "50055")
	viper.SetDefault("grpc.max_recv_msg_size", 4194304)
	viper.SetDefault("grpc.max_send_msg_size", 4194304)
	viper.SetDefault("grpc.timeout", "15s")

	viper.SetDefault("mongo.uri", "mongodb://localhost:27017")
	viper.SetDefault("mongo.database", "news_service_db")
	viper.SetDefault("mongo.connect_timeout", "10s")
	viper.SetDefault("mongo.username", "")
	viper.SetDefault("mongo.password", "")
	viper.SetDefault("mongo.min_pool_size", 0)
	viper.SetDefault("mongo.max_pool_size", 50)

	viper.SetDefault("nats.url", "nats://localhost:4222")
	viper.SetDefault("nats.connect_timeout", "5s")

	viper.SetDefault("redis.address", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	viper.SetDefault("user_service_address", "localhost:50051")

	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println(".env file not found by Viper, relying on actual environment variables or defaults.")
		} else {
			log.Printf("Error reading .env file with Viper: %s\n", err)
		}
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		if fi, _ := os.Stat(path); !fi.IsDir() {
			viper.SetConfigFile(path)
		} else {
			viper.AddConfigPath(path)
			viper.SetConfigName("config")
		}
		viper.SetConfigType("yaml")
		if err := viper.MergeInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				log.Println("YAML config file specified by path not found, relying on .env or defaults.")
			} else {
				return nil, fmt.Errorf("error merging YAML config file: %w", err)
			}
		}
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		if err := viper.MergeInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				log.Println("Default config.yaml not found in current directory, relying on .env or defaults.")
			} else {
				log.Printf("Error reading default config.yaml: %s\n", err)
			}
		}
	}

	viper.SetEnvPrefix("NEWS")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct (initial unmarshal): %v", err)
	}

	smtpHost := os.Getenv("NEWS_SMTP_HOST")
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}
	cfg.SMTP.Host = smtpHost

	smtpPortStr := os.Getenv("NEWS_SMTP_PORT")
	smtpPort := 587
	if smtpPortStr != "" {
		parsedPort, err := strconv.Atoi(smtpPortStr)
		if err == nil {
			smtpPort = parsedPort
		} else {
			log.Printf("Warning: Could not parse NEWS_SMTP_PORT ('%s') to int. Using default %d.\n", smtpPortStr, smtpPort)
		}
	}
	cfg.SMTP.Port = smtpPort

	cfg.SMTP.Username = os.Getenv("NEWS_SMTP_USERNAME")
	cfg.SMTP.Password = os.Getenv("NEWS_SMTP_PASSWORD")
	cfg.SMTP.SenderEmail = os.Getenv("NEWS_SMTP_SENDER_EMAIL")

	if cfg.UserServiceAddress == "" {
		cfg.UserServiceAddress = os.Getenv("NEWS_USER_SERVICE_ADDRESS")
		if cfg.UserServiceAddress == "" {
			cfg.UserServiceAddress = "localhost:50051"
		}
	}

	return &cfg, nil
}
