package main

import (
	"context"
	"log"
	"os"

	mongoAdapter "github.com/Abdurahmanit/GroupProject/news-service/internal/adapter/mongo"
	"github.com/Abdurahmanit/GroupProject/news-service/internal/config"
	"github.com/Abdurahmanit/GroupProject/news-service/internal/usecase"
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

	mongoClient, err := mongoAdapter.NewMongoDBConnection(&cfg.Mongo)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	defer func() {
		if err = mongoClient.Disconnect(context.TODO()); err != nil {
			logger.Error("Failed to disconnect MongoDB", zap.Error(err))
		} else {
			logger.Info("MongoDB connection closed.")
		}
	}()
	logger.Info("Successfully connected to MongoDB!")

	newsRepo := mongoAdapter.NewNewsMongoRepository(mongoClient, cfg.Mongo.Database)
	logger.Info("News repository initialized")

	newsUC := usecase.NewNewsUseCase(newsRepo)
	_ = newsUC
	logger.Info("News use case initialized")

	logger.Info("News Service setup complete. Ready to start gRPC server.", zap.String("port", cfg.GRPC.Port))
}
