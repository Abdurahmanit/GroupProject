package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	listingserviceclient "github.com/Abdurahmanit/GroupProject/order-service/internal/adapter/client"
	mongoadapter "github.com/Abdurahmanit/GroupProject/order-service/internal/adapter/mongo"
	natsadapter "github.com/Abdurahmanit/GroupProject/order-service/internal/adapter/nats"
	redisadapter "github.com/Abdurahmanit/GroupProject/order-service/internal/adapter/redis"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/app/config"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/platform/logger"
	grpcport "github.com/Abdurahmanit/GroupProject/order-service/internal/port/grpc"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/repository"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/service"

	listingpb "github.com/Abdurahmanit/GroupProject/listing-service/genproto/listing_service"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
)

type App struct {
	cfg                  *config.Config
	log                  logger.Logger
	server               *grpcport.Server
	orderRepo            repository.OrderRepository
	cartRepo             repository.CartRepository
	productCacheRepo     repository.ProductDetailCache
	msgPublisher         natsadapter.MessagePublisher
	listingServiceClient listingpb.ListingServiceClient
	cartService          service.CartService
	orderService         service.OrderService
	receiptService       service.ReceiptService
	mongoClient          *mongo.Client
	redisClient          *redis.Client
	natsConn             *nats.Conn
	listingServiceConn   *grpc.ClientConn
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
		mongoClient.Disconnect(ctx)
		return nil, fmt.Errorf("failed to initialize Redis client: %w", err)
	}
	appLogger.Info("Redis client initialized successfully")

	appLogger.Info("Initializing NATS connection...")
	natsConn, err := natsadapter.NewConnection(cfg.NATS)
	if err != nil {
		appLogger.Errorf("Failed to initialize NATS connection: %v", err)
		mongoClient.Disconnect(ctx)
		redisClient.Close()
		return nil, fmt.Errorf("failed to initialize NATS connection: %w", err)
	}
	appLogger.Info("NATS connection initialized successfully")

	msgPublisher, err := natsadapter.NewNATSPublisher(natsConn)
	if err != nil {
		appLogger.Errorf("Failed to initialize NATS publisher: %v", err)
		natsConn.Close()
		mongoClient.Disconnect(ctx)
		redisClient.Close()
		return nil, fmt.Errorf("failed to initialize NATS publisher: %w", err)
	}
	appLogger.Info("NATS MessagePublisher initialized")

	appLogger.Info("Initializing ListingService gRPC client...")
	listingServiceClientCfg := listingserviceclient.ListingServiceClientConfig{
		Address: cfg.Services.ListingService.Address,
	}
	listingServiceCl, listingServiceConn, err := listingserviceclient.NewListingServiceClient(listingServiceClientCfg)
	if err != nil {
		appLogger.Errorf("Failed to initialize ListingService client: %v", err)
		natsConn.Close()
		mongoClient.Disconnect(ctx)
		redisClient.Close()
		return nil, fmt.Errorf("failed to initialize ListingService client: %w", err)
	}
	appLogger.Info("ListingService gRPC client initialized successfully")

	orderRepo := mongoadapter.NewOrderRepository(mongoClient, cfg.MongoDB)
	appLogger.Info("OrderRepository initialized")
	cartRepo := redisadapter.NewCartRepository(redisClient)
	appLogger.Info("CartRepository initialized")
	productCache := redisadapter.NewProductDetailCacheRepository(redisClient)
	appLogger.Info("ProductDetailCacheRepository initialized")

	cartServiceCfg := service.CartServiceConfig{
		CartTTL:         cfg.Cart.TTL,
		ProductCacheTTL: cfg.ProductCache.TTL,
	}
	cartSvc := service.NewCartService(cartRepo, productCache, listingServiceCl, appLogger, cartServiceCfg)
	appLogger.Info("CartService initialized")

	orderSvc := service.NewOrderService(orderRepo, cartSvc, listingServiceCl, msgPublisher, appLogger)
	appLogger.Info("OrderService initialized")

	receiptSvc := service.NewReceiptService(orderRepo, appLogger)
	appLogger.Info("ReceiptService initialized")

	orderGRPCHandler := grpcport.NewOrderGRPCHandler(cartSvc, orderSvc, receiptSvc, appLogger)
	appLogger.Info("OrderGRPCHandler initialized")

	grpcSrv := grpcport.NewServer(
		appLogger,
		cfg.GRPCServer.Port,
		cfg.GRPCServer.TimeoutGraceful,
		cfg.GRPCServer.MaxConnectionIdle,
		orderGRPCHandler,
	)
	appLogger.Info("gRPC server instance created with OrderService handler")

	application := &App{
		cfg:                  cfg,
		log:                  appLogger,
		server:               grpcSrv,
		orderRepo:            orderRepo,
		cartRepo:             cartRepo,
		productCacheRepo:     productCache,
		msgPublisher:         msgPublisher,
		listingServiceClient: listingServiceCl,
		cartService:          cartSvc,
		orderService:         orderSvc,
		receiptService:       receiptSvc,
		mongoClient:          mongoClient,
		redisClient:          redisClient,
		natsConn:             natsConn,
		listingServiceConn:   listingServiceConn,
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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.GRPCServer.TimeoutGraceful+10*time.Second)
	defer cancel()

	if err := a.server.Stop(shutdownCtx); err != nil {
		a.log.Errorf("Error during gRPC server graceful shutdown: %v", err)
	} else {
		a.log.Info("gRPC server stopped successfully")
	}

	a.log.Info("Closing infrastructure connections...")

	if a.listingServiceConn != nil {
		a.log.Info("Closing ListingService gRPC client connection...")
		if err := a.listingServiceConn.Close(); err != nil {
			a.log.Errorf("Error closing ListingService gRPC client connection: %v", err)
		} else {
			a.log.Info("ListingService gRPC client connection closed successfully")
		}
	}

	if a.natsConn != nil {
		if !a.natsConn.IsClosed() {
			a.log.Info("Draining NATS connection...")
			if err := a.natsConn.Drain(); err != nil {
				a.log.Errorf("Error draining NATS connection: %v", err)
			} else {
				a.log.Info("NATS connection drained successfully")
			}
		}
		a.natsConn.Close()
		a.log.Info("NATS connection closed")
	}

	if a.redisClient != nil {
		if err := a.redisClient.Close(); err != nil {
			a.log.Errorf("Error closing Redis client: %v", err)
		} else {
			a.log.Info("Redis client closed successfully")
		}
	}

	if a.mongoClient != nil {
		if err := a.mongoClient.Disconnect(shutdownCtx); err != nil {
			a.log.Errorf("Error disconnecting from MongoDB: %v", err)
		} else {
			a.log.Info("MongoDB connection closed successfully")
		}
	}

	a.log.Info("Application shut down successfully")
}
