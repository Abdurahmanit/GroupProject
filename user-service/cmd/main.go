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
	"time"

	"github.com/Abdurahmanit/GroupProject/user-service/internal/adapter"
	"github.com/Abdurahmanit/GroupProject/user-service/internal/config"
	"github.com/Abdurahmanit/GroupProject/user-service/internal/mailer"
	"github.com/Abdurahmanit/GroupProject/user-service/internal/repository"
	"github.com/Abdurahmanit/GroupProject/user-service/internal/usecase"
	user "github.com/Abdurahmanit/GroupProject/user-service/proto"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("INFO: .env file not found or error loading. Error:", err)
	}

	mongoURIFromEnv := os.Getenv("MONGO_URI")
	log.Printf("DEBUG: MONGO_URI from os.Getenv() after godotenv.Load(): [%s]\n", mongoURIFromEnv)

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("CRITICAL: Can't initialize zap logger: %v", err)
	}
	defer logger.Sync()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config with Viper", zap.Error(err))
	}

	// Configuration validation (essential fields)
	if cfg.MongoURI == "" {
		logger.Fatal("FATAL: cfg.MongoURI is empty.")
	}
	if cfg.RedisAddr == "" {
		logger.Fatal("FATAL: cfg.RedisAddr is empty.")
	}
	if cfg.JWTSecret == "" {
		logger.Warn("WARNING: cfg.JWTSecret is empty. This is insecure.")
	}
	if cfg.Port == 0 {
		logger.Warn("WARNING: cfg.Port is 0. Defaulting to 50051.", zap.Int("current_cfg_port", cfg.Port))
		cfg.Port = 50051 // Ensure a default port if not set
	}

	// Mailer configuration validation based on MAILER_TYPE
	var mailerService mailer.Mailer // Use the interface type
	logger.Info("Configured MAILER_TYPE", zap.String("type", cfg.MailerType))

	if cfg.MailerType == "smtp" {
		logger.Info("Initializing SMTP Mailer Service")
		if cfg.SMTPHost == "" || cfg.SMTPPort == 0 || cfg.SMTPUsername == "" || cfg.SMTPPassword == "" || cfg.SMTPFromEmail == "" {
			logger.Fatal("FATAL: SMTP configuration is incomplete (HOST, PORT, USERNAME, PASSWORD, FROM_EMAIL are required for SMTP mailer type).")
		}
		mailerService = mailer.NewSMTPMailerService(
			cfg.SMTPHost,
			cfg.SMTPPort,
			cfg.SMTPUsername,
			cfg.SMTPPassword,
			cfg.SMTPFromEmail,
			cfg.SMTPSenderName, // This can be empty if not set, mailer will handle
			logger,
		)
	} else if cfg.MailerType == "mailersend" {
		logger.Info("Initializing MailerSend API Service")
		if cfg.MailerSendAPIKey == "" || cfg.MailerSendFromEmail == "" {
			logger.Fatal("FATAL: MailerSend configuration is incomplete (API_KEY, FROM_EMAIL are required for mailersend mailer type).")
		}
		mailerService = mailer.NewMailerSendService(
			cfg.MailerSendAPIKey,
			cfg.MailerSendFromEmail,
			cfg.MailerSendFromName, // This can be empty if not set, mailer will handle
			logger,
		)
	} else {
		logger.Fatal("Invalid MAILER_TYPE specified in configuration. Choose 'smtp' or 'mailersend'.")
	}

	// Connect to MongoDB
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.String("mongoURI_used", cfg.MongoURI), zap.Error(err))
	}
	defer func() {
		logger.Info("Disconnecting MongoDB...")
		if errDisconnect := mongoClient.Disconnect(context.Background()); errDisconnect != nil {
			logger.Error("Failed to disconnect MongoDB", zap.Error(errDisconnect))
		}
	}()
	ctxPingMongo, cancelPingMongo := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPingMongo()
	if err := mongoClient.Ping(ctxPingMongo, nil); err != nil {
		logger.Fatal("Failed to ping MongoDB", zap.String("mongoURI_used", cfg.MongoURI), zap.Error(err))
	}
	logger.Info("Successfully connected to MongoDB", zap.String("mongoURI_used", cfg.MongoURI))
	db := mongoClient.Database("bicycle_shop")

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	ctxPingRedis, cancelPingRedis := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPingRedis()
	_, err = redisClient.Ping(ctxPingRedis).Result()
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.String("redisAddr_used", cfg.RedisAddr), zap.Error(err))
	}
	logger.Info("Successfully connected to Redis", zap.String("redisAddr_used", cfg.RedisAddr))
	defer func() {
		logger.Info("Closing Redis connection...")
		if errClose := redisClient.Close(); errClose != nil {
			logger.Error("Failed to close Redis connection", zap.Error(errClose))
		}
	}()

	// Initialize components
	userRepo := repository.NewUserRepository(db, redisClient, logger)
	userUsecase := usecase.NewUserUsecase(userRepo, mailerService, cfg.JWTSecret, logger) // Pass the chosen mailerService
	userGRPCHandler := adapter.NewUserHandler(userUsecase, logger)

	// Start gRPC server
	address := fmt.Sprintf(":%d", cfg.Port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		logger.Fatal("Failed to listen on address", zap.String("address", address), zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	user.RegisterUserServiceServer(grpcServer, userGRPCHandler)
	logger.Info("Starting User Service gRPC server", zap.String("address", address))

	go func() {
		if errServe := grpcServer.Serve(lis); errServe != nil && !errors.Is(errServe, grpc.ErrServerStopped) {
			logger.Fatal("Failed to serve gRPC", zap.Error(errServe))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	logger.Info("User Service stopped gracefully.")
}
