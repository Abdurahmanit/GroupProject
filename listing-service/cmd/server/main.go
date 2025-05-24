package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time" // Для таймаута при закрытии трейсера

	grpcAdapter "github.com/Abdurahmanit/GroupProject/listing-service/internal/adapter/grpc"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/adapter/messaging/nats"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/adapter/repository/mongodb"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/adapter/storage/s3"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/adapter/repository/cache"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/config"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/platform/logger"   // <--- ПУТЬ К ТВОЕМУ ЛОГГЕРУ
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/platform/tracer"   // <--- ПУТЬ К ТВОЕМУ ТРЕЙСЕРУ
	pb "github.com/Abdurahmanit/GroupProject/listing-service/genproto/listing_service"


	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Инициализация логгера в первую очередь
	appLogger := logger.NewLogger() // Используем твой конструктор
	appLogger.Info("Application starting...") // Первое сообщение через кастомный логгер

	// Инициализация трейсера
	tp := tracer.InitTracer()
	defer func() {
		appLogger.Info("Shutting down tracer provider...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			appLogger.Error("Failed to shutdown tracer provider", "error", err)
		} else {
			appLogger.Info("Tracer provider shut down successfully.")
		}
	}()


	// Load configuration
	cfg, err := config.Load() // config.Load может использовать os.Getenv, которые ты настраиваешь
	if err != nil {
		appLogger.Error("Failed to load config", "error", err)
		os.Exit(1) // Завершаем, если конфиг не загружен
	}
	appLogger.Info("Configuration loaded successfully.")

	// Connect to MongoDB
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		appLogger.Error("Failed to connect to MongoDB", "uri", cfg.MongoURI, "error", err)
		os.Exit(1)
	}
	defer func() {
		appLogger.Info("Disconnecting from MongoDB...")
		if err = mongoClient.Disconnect(context.Background()); err != nil {
			appLogger.Error("Error disconnecting from MongoDB", "error", err)
		} else {
			appLogger.Info("Disconnected from MongoDB successfully.")
		}
	}()
	db := mongoClient.Database("bike_store")
	appLogger.Info("Successfully connected to MongoDB.")

	// Initialize repositories
	listingRepo := mongodb.NewListingRepository(db, appLogger)     // Передай логгер, если репозиторий его использует
	favoriteRepo := mongodb.NewFavoriteRepository(db, appLogger) // Аналогично
	appLogger.Info("Repositories initialized.")

	// Initialize ListingCache (Redis)
	listingCache, err := cache.NewListingCache(cfg.RedisAddress)
	if err != nil {
		appLogger.Error("Failed to initialize listing cache", "address", cfg.RedisAddress, "error", err)
		os.Exit(1)
	}
	defer func() {
		appLogger.Info("Closing Redis client...")
		if err := listingCache.CloseClient(context.Background()); err != nil { // Убедись, что CloseClient есть
			appLogger.Error("Error closing Redis client", "error", err)
		} else {
			appLogger.Info("Redis client closed successfully.")
		}
	}()
	appLogger.Info("ListingCache (Redis) initialized successfully.")

	// Initialize storage (MinIO/S3)
	storageClient, err := s3.NewS3Storage(cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinIOBucket, cfg.MinIOUseSSL, appLogger) // <--- ПЕРЕДАЕМ ЛОГГЕР В S3
	if err != nil {
		appLogger.Error("Failed to initialize S3 storage", "endpoint", cfg.MinIOEndpoint, "error", err)
		os.Exit(1)
	}
	appLogger.Info("S3 storage initialized.")

	// Initialize NATS publisher
	natsPublisher, err := nats.NewPublisher(cfg.NATSURL, appLogger) // <--- ПЕРЕДАЕМ ЛОГГЕР В NATS
	if err != nil {
		appLogger.Error("Failed to initialize NATS publisher", "url", cfg.NATSURL, "error", err)
		os.Exit(1)
	}
	defer func() {
		appLogger.Info("Closing NATS publisher...")
		natsPublisher.Close() // Предполагается, что NATS клиент имеет метод Close
		appLogger.Info("NATS publisher closed.")
	}()
	appLogger.Info("NATS publisher initialized.")

	// Set up gRPC server
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		appLogger.Error("Failed to listen for gRPC", "port", cfg.GRPCPort, "error", err)
		os.Exit(1)
	}

	// grpcAdapter.NewGRPCServer() вероятно создает *grpc.Server и возвращает его и функцию cleanup.
	// cleanup обычно вызывает server.GracefulStop() или server.Stop()
	// Можно также передать appLogger в grpcAdapter.NewGRPCServer(), если там нужны логи
	grpcSrv, cleanup := grpcAdapter.NewGRPCServer(appLogger, cfg.JWTSecret) // <--- ПЕРЕДАЕМ ЛОГГЕР В GRPC SERVER ADAPTER

	// Передаем appLogger в Handler
	handler := grpcAdapter.NewHandler(listingRepo, favoriteRepo, storageClient, natsPublisher, listingCache, appLogger) // <--- ЛОГГЕР ПЕРЕДАН В GRPC HANDLER
	pb.RegisterListingServiceServer(grpcSrv, handler)

	// Graceful Shutdown
	go func() {
		appLogger.Info("Starting gRPC server", "port", cfg.GRPCPort)
		if err := grpcSrv.Serve(lis); err != nil {
			// Эта ошибка возникает при штатном завершении через GracefulStop, поэтому не Fatal
			appLogger.Error("gRPC server Serve error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down gRPC server...")
	cleanup() // Вызываем cleanup от gRPC сервера (например, grpcSrv.GracefulStop())
	appLogger.Info("gRPC server stopped.")

	appLogger.Info("Application shutting down...")
	// Остальные defer'ы (mongo, redis, nats, tracer) будут выполнены после выхода из main
}