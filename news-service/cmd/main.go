package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	redisAdapter "github.com/Abdurahmanit/GroupProject/news-service/internal/adapter/cache/redis"
	mongoAdapter "github.com/Abdurahmanit/GroupProject/news-service/internal/adapter/mongo"
	natsAdapter "github.com/Abdurahmanit/GroupProject/news-service/internal/adapter/nats"
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
		zap.String("nats_url", cfg.NATS.URL),
		zap.String("redis_address", cfg.Redis.Address),
	)

	mongoClient, err := mongoAdapter.NewMongoDBConnection(&cfg.Mongo)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	defer func() {
		logger.Info("Attempting to disconnect MongoDB...")
		if err = mongoClient.Disconnect(context.TODO()); err != nil {
			logger.Error("Failed to disconnect MongoDB", zap.Error(err))
		} else {
			logger.Info("MongoDB connection closed.")
		}
	}()
	logger.Info("Successfully connected to MongoDB!")

	natsPublisher, err := natsAdapter.NewNATSPublisher(&cfg.NATS, logger)
	if err != nil {
		logger.Fatal("Failed to connect to NATS", zap.Error(err))
	}
	defer func() {
		logger.Info("Attempting to close NATS publisher connection...")
		natsPublisher.Close()
	}()
	logger.Info("Successfully connected to NATS!")

	redisClient, err := redisAdapter.NewRedisClient(&cfg.Redis, logger)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer func() {
		logger.Info("Attempting to close Redis client connection...")
		if err := redisClient.Close(); err != nil {
			logger.Error("Failed to close Redis client connection", zap.Error(err))
		} else {
			logger.Info("Redis client connection closed.")
		}
	}()

	newsRepo := mongoAdapter.NewNewsMongoRepository(mongoClient, cfg.Mongo.Database)
	commentRepo := mongoAdapter.NewCommentMongoRepository(mongoClient, cfg.Mongo.Database)
	likeRepo := mongoAdapter.NewLikeMongoRepository(mongoClient, cfg.Mongo.Database)

	cacheRepo := redisAdapter.NewRedisCacheRepository(redisClient, logger)
	logger.Info("Repositories (DB & Cache) initialized")

	newsUC := usecase.NewNewsUseCase(newsRepo, natsPublisher, cacheRepo, logger)
	commentUC := usecase.NewCommentUseCase(commentRepo, newsRepo)
	likeUC := usecase.NewLikeUseCase(likeRepo, newsRepo, commentRepo)

	logger.Info("Use cases initialized")

	newsGRPCHandler := grpcPort.NewNewsHandler(newsUC, commentUC, likeUC)
	grpcServer := grpcPort.NewServer(&cfg.GRPC, logger, newsGRPCHandler)

	logger.Info("Starting gRPC server...", zap.String("port", cfg.GRPC.Port))
	go func() {
		if err := grpcServer.Run(); err != nil {
			logger.Fatal("gRPC server failed to run", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	logger.Info("Received shutdown signal", zap.String("signal", sig.String()))

	logger.Info("Shutting down gRPC server (will stop on its own after listener closes or by OS signal)...")

	logger.Info("News Service shut down gracefully.")
}
