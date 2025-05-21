package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	mongoAdapter "github.com/Abdurahmanit/GroupProject/news-service/internal/adapter/mongo"
	"github.com/Abdurahmanit/GroupProject/news-service/internal/config"
	grpcPort "github.com/Abdurahmanit/GroupProject/news-service/internal/port/grpc"
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
	defer func() {
		_ = logger.Sync()
	}()

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
	commentRepo := mongoAdapter.NewCommentMongoRepository(mongoClient, cfg.Mongo.Database)
	likeRepo := mongoAdapter.NewLikeMongoRepository(mongoClient, cfg.Mongo.Database)
	logger.Info("Repositories initialized")

	newsUC := usecase.NewNewsUseCase(newsRepo)
	commentUC := usecase.NewCommentUseCase(commentRepo, newsRepo)
	likeUC := usecase.NewLikeUseCase(likeRepo, newsRepo, commentRepo)
	_ = commentUC
	_ = likeUC
	logger.Info("Use cases initialized")

	newsGRPCHandler := grpcPort.NewNewsHandler(newsUC)

	grpcServer := grpcPort.NewServer(&cfg.GRPC, logger, newsGRPCHandler)

	logger.Info("Starting gRPC server...", zap.String("port", cfg.GRPC.Port))

	go func() {
		if err := grpcServer.Run(); err != nil {
			logger.Fatal("gRPC server failed to run", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down gRPC server...")
	logger.Info("News Service shut down.")
}
