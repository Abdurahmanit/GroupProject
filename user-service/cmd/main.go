package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/Abdurahmanit/GroupProject/user-service/internal/adapter"
	"github.com/Abdurahmanit/GroupProject/user-service/internal/config"
	"github.com/Abdurahmanit/GroupProject/user-service/internal/repository"
	"github.com/Abdurahmanit/GroupProject/user-service/internal/usecase"
	user "github.com/Abdurahmanit/GroupProject/user-service/proto"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Connect to MongoDB
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	// Defer disconnect in a separate goroutine or ensure it's called on shutdown
	// For simplicity here, direct defer. In production, handle signals for graceful shutdown.
	defer func() {
		if err = mongoClient.Disconnect(context.Background()); err != nil {
			logger.Error("Failed to disconnect MongoDB", zap.Error(err))
		}
	}()
	db := mongoClient.Database("bicycle_shop") // Or from cfg

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	userRepo := repository.NewUserRepository(db, redisClient)
	userUsecase := usecase.NewUserUsecase(userRepo, cfg.JWTSecret)
	userHandler := adapter.NewUserHandler(userUsecase, logger)

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err), zap.Int("port", cfg.Port))
	}

	grpcServer := grpc.NewServer()
	user.RegisterUserServiceServer(grpcServer, userHandler)

	logger.Info("Starting User Service", zap.Int("port", cfg.Port))
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("Failed to serve gRPC", zap.Error(err))
	}
}
