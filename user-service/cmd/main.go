package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Abdurahmanit/GroupProject/user-service/internal/adapter"
	"github.com/Abdurahmanit/GroupProject/user-service/internal/config"
	"github.com/Abdurahmanit/GroupProject/user-service/internal/mailer" // Import mailer
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
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("failed to sync logger: %v\n", err)
		}
	}()

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
	defer func() {
		logger.Info("Disconnecting MongoDB...")
		if err = mongoClient.Disconnect(context.Background()); err != nil {
			logger.Error("Failed to disconnect MongoDB", zap.Error(err))
		}
	}()
	db := mongoClient.Database("bicycle_shop") // Consider making DB name configurable

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer func() {
		logger.Info("Closing Redis connection...")
		if err := redisClient.Close(); err != nil {
			logger.Error("Failed to close Redis connection", zap.Error(err))
		}
	}()

	// Initialize Mailer Service
	mailerService := mailer.NewMailerSendService(cfg.MailerSendAPIKey, cfg.MailerSendFromEmail, cfg.MailerSendFromName, logger)

	// Initialize components
	userRepo := repository.NewUserRepository(db, redisClient, logger)
	userUsecase := usecase.NewUserUsecase(userRepo, mailerService, cfg.JWTSecret, logger)
	userGRPCHandler := adapter.NewUserHandler(userUsecase, logger)

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err), zap.Int("port", cfg.Port))
	}

	grpcServer := grpc.NewServer()
	user.RegisterUserServiceServer(grpcServer, userGRPCHandler)

	logger.Info("Starting User Service", zap.Int("port", cfg.Port))

	// Graceful shutdown
	go func() {
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			logger.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	logger.Info("User Service stopped.")
}
