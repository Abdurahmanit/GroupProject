package config

import (
	"log"
	"os"
	"strconv" // Для конвертации строки в bool

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI       string
	NATSURL        string
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool   // <--- ДОБАВЛЕНО
	GRPCPort       string
	RedisAddress   string
	JWTSecret      string // <--- ДОБАВЛЕНО
	// AWSRegion      string // Добавь, если используешь AWS S3 SDK и нужен регион
}

func Load() (*Config, error) {
	// Загружаем .env файл, если он есть. Ошибку здесь можно игнорировать,
	// если переменные окружения являются основным источником.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading .env, relying on environment variables")
	}

	minioUseSSLStr := getEnv("MINIO_USE_SSL", "false") // По умолчанию false
	minioUseSSL, err := strconv.ParseBool(minioUseSSLStr)
	if err != nil {
		log.Printf("Warning: Invalid MINIO_USE_SSL value '%s', defaulting to false. Error: %v", minioUseSSLStr, err)
		minioUseSSL = false // Безопасное значение по умолчанию при ошибке парсинга
	}

	cfg := &Config{
		MongoURI:       getEnv("MONGO_URI", "mongodb://localhost:27017"),
		NATSURL:        getEnv("NATS_URL", "nats://localhost:4222"),
		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"), // Для MinIO эндпоинт обычно без http(s)://
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucket:    getEnv("MINIO_BUCKET", "listings-photos"),
		MinIOUseSSL:    minioUseSSL, // <--- УСТАНОВЛЕНО
		GRPCPort:       getEnv("GRPC_PORT", "50052"), // Убедись, что этот порт не конфликтует с другими сервисами
		RedisAddress:   getEnv("REDIS_ADDRESS", "localhost:6379"),
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key"), // <--- УСТАНОВЛЕНО (ВАЖНО: измени дефолтное значение)
		// AWSRegion:      getEnv("AWS_REGION", "us-east-1"), // Если используешь AWS S3 SDK
	}

	// Валидация критичных полей, например JWTSecret
	if cfg.JWTSecret == "your-secret-key" {
		log.Println("Warning: JWT_SECRET is set to its default insecure value. Please set a strong secret in your environment or .env file.")
	}
	if cfg.JWTSecret == "" {
	    // Можно завершить приложение, если JWT_SECRET обязателен и пуст
	    log.Fatal("FATAL: JWT_SECRET is not set. This is required for security.")
	}


	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Printf("Environment variable %s not set, using fallback: %s", key, fallback)
	return fallback
}