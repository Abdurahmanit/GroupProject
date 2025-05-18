package main

import (
	"context" // Добавлен импорт
	"log"
	"os"

	// Путь к вашему пакету mongo (если он отличается, поправь)
	"github.com/Abdurahmanit/GroupProject/news-service/internal/adapter/mongo"
	"github.com/Abdurahmanit/GroupProject/news-service/internal/config"
	"go.uber.org/zap"
)

func main() {
	configPath := "config.yaml"
	if cp := os.Getenv("CONFIG_PATH"); cp != "" {
		configPath = cp
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	zapConfig := zap.NewProductionConfig()
	logger, err := zapConfig.Build()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Configuration loaded successfully!",
		zap.String("grpc_port", cfg.GRPC.Port),
		zap.String("mongo_uri", cfg.Mongo.URI),
		zap.String("mongo_database", cfg.Mongo.Database),
	)

	// Инициализация MongoDB
	mongoClient, err := mongo.NewMongoDBConnection(&cfg.Mongo)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	// Важно закрывать соединение при завершении работы приложения
	defer func() {
		if err = mongoClient.Disconnect(context.TODO()); err != nil { // Используем context.TODO() для Disconnect
			logger.Error("Failed to disconnect MongoDB", zap.Error(err))
		} else {
			logger.Info("MongoDB connection closed.")
		}
	}()
	logger.Info("Successfully connected to MongoDB!")

	logger.Info("News Service starting", zap.String("port", cfg.GRPC.Port))

	// Здесь будет код для ожидания сигнала завершения и graceful shutdown,
	// а пока что приложение просто завершится после вывода логов.
}
