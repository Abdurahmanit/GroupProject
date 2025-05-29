package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	mongoadapter "github.com/Abdurahmanit/GroupProject/order-service/internal/adapter/mongo"
	redisadapter "github.com/Abdurahmanit/GroupProject/order-service/internal/adapter/redis"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/app/config"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/platform/logger"
	grpcserver "github.com/Abdurahmanit/GroupProject/order-service/internal/port/grpc"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/repository"
	orderservicepb "github.com/Abdurahmanit/GroupProject/order-service/proto/service"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type App struct {
	cfg         *config.Config
	log         logger.Logger
	server      *grpcserver.Server
	orderRepo   repository.OrderRepository
	cartRepo    repository.CartRepository
	mongoClient *mongo.Client
	redisClient *redis.Client
}

func New(cfg *config.Config) (*App, error) {
	ctx := context.Background()

	logCfg := logger.ZapLoggerConfig{
		Level:      cfg.Logger.Level,
		Encoding:   cfg.Logger.Encoding,
		TimeFormat: cfg.Logger.TimeFormat,
	}
	appLogger, err := logger.NewZapLogger(logCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	appLogger.Info("Logger initialized")
	appLogger.Infof("Configuration loaded: Env=%s, GRPC Port: %s", cfg.Env, cfg.GRPCServer.Port)

	appLogger.Info("Initializing MongoDB client...")
	mongoClient, err := mongoadapter.NewClient(ctx, cfg.MongoDB)
	if err != nil {
		appLogger.Errorf("Failed to initialize MongoDB client: %v", err)
		return nil, fmt.Errorf("failed to initialize MongoDB client: %w", err)
	}
	appLogger.Info("MongoDB client initialized successfully")

	appLogger.Info("Initializing Redis client...")
	redisClient, err := redisadapter.NewClient(ctx, cfg.Redis)
	if err != nil {
		appLogger.Errorf("Failed to initialize Redis client: %v", err)
		return nil, fmt.Errorf("failed to initialize Redis client: %w", err)
	}
	appLogger.Info("Redis client initialized successfully")

	orderRepo := mongoadapter.NewOrderRepository(mongoClient, cfg.MongoDB)
	appLogger.Info("OrderRepository initialized")
	cartRepo := redisadapter.NewCartRepository(redisClient)
	appLogger.Info("CartRepository initialized")

	var orderServiceImplementation orderservicepb.OrderServiceServer

	grpcSrv := grpcserver.NewServer(
		appLogger,
		cfg.GRPCServer.Port,
		cfg.GRPCServer.TimeoutGraceful,
		cfg.GRPCServer.MaxConnectionIdle,
		orderServiceImplementation,
	)
	appLogger.Info("gRPC server instance created")

	application := &App{
		cfg:         cfg,
		log:         appLogger,
		server:      grpcSrv,
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		mongoClient: mongoClient,
		redisClient: redisClient,
	}

	return application, nil
}

func (a *App) Run() {
	a.log.Info("Starting application components...")

	go func() {
		if err := a.server.Start(); err != nil {
			a.log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()
	a.log.Info("gRPC server started in a goroutine")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	receivedSignal := <-quit
	a.log.Infof("Received shutdown signal: %v. Shutting down application...", receivedSignal)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.GRPCServer.TimeoutGraceful+5*time.Second)
	defer cancel()

	if err := a.server.Stop(shutdownCtx); err != nil {
		a.log.Errorf("Error during gRPC server graceful shutdown: %v", err)
	} else {
		a.log.Info("gRPC server stopped successfully")
	}

	a.log.Info("Closing database connections...")

	if a.mongoClient != nil {
		if err := a.mongoClient.Disconnect(shutdownCtx); err != nil {
			a.log.Errorf("Error disconnecting from MongoDB: %v", err)
		} else {
			a.log.Info("MongoDB connection closed successfully")
		}
	}

	if a.redisClient != nil {
		if err := a.redisClient.Close(); err != nil {
			a.log.Errorf("Error closing Redis client: %v", err)
		} else {
			a.log.Info("Redis client closed successfully")
		}
	}

	a.log.Info("Application shut down successfully")
}
