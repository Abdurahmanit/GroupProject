package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcAdapter "github.com/Abdurahmanit/GroupProject/review-service/internal/adapter/grpc"
	natsAdapter "github.com/Abdurahmanit/GroupProject/review-service/internal/adapter/messaging/nats"
	mongoRepo "github.com/Abdurahmanit/GroupProject/review-service/internal/adapter/repository/mongodb"

	"github.com/Abdurahmanit/GroupProject/review-service/internal/config"
	"github.com/Abdurahmanit/GroupProject/review-service/internal/platform/logger"
	"github.com/Abdurahmanit/GroupProject/review-service/internal/platform/metrics"
	"github.com/Abdurahmanit/GroupProject/review-service/internal/platform/tracer"
	"github.com/Abdurahmanit/GroupProject/review-service/internal/usecase"

	pb "github.com/Abdurahmanit/GroupProject/review-service"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	serviceName = "review-service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("INFO: .env file not found or error loading: %v. Relying on OS environment variables.\n", err)
	}

	// 1. Initialize Logger
	appLogger := logger.NewLogger()
	appLogger.Info("Application starting...", zap.String("service_name", serviceName))

	// 2. Load Configuration
	cfg, err := config.LoadConfig(appLogger) // Pass logger to config loading
	if err != nil {
		appLogger.Fatal("Failed to load configuration", zap.Error(err))
	}
	appLogger.Info("Configuration loaded successfully",
		zap.String("grpc_port", cfg.GRPCPort),
		zap.Bool("mongo_uri_set", cfg.MongoURI != ""), // Corrected: Use zap.Bool for boolean
		zap.String("nats_url", cfg.NATSURL),
		zap.String("prometheus_port", cfg.PrometheusMetricsPort),
	)

	// 3. Initialize OpenTelemetry Tracer
	var tp *sdktrace.TracerProvider
	if cfg.OTExporterOTLPEndpoint != "" {
		tp = tracer.InitTracer(serviceName, cfg.OTExporterOTLPEndpoint, appLogger)
		defer func() {
			appLogger.Info("Shutting down tracer provider...")
			ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelShutdown()
			if err := tp.Shutdown(ctxShutdown); err != nil {
				appLogger.Error("Failed to shutdown tracer provider", zap.Error(err))
			} else {
				appLogger.Info("Tracer provider shut down successfully.")
			}
		}()
		appLogger.Info("OpenTelemetry Tracer initialized.")
	} else {
		appLogger.Info("OpenTelemetry Tracer not initialized (OTEL_EXPORTER_OTLP_ENDPOINT not set).")
	}

	// 4. Connect to MongoDB
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		appLogger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	defer func() {
		appLogger.Info("Disconnecting from MongoDB...")
		if err = mongoClient.Disconnect(context.Background()); err != nil {
			appLogger.Error("Error disconnecting from MongoDB", zap.Error(err))
		} else {
			appLogger.Info("Disconnected from MongoDB successfully.")
		}
	}()
	// Ping MongoDB to ensure connection
	ctxPingMongo, cancelPingMongo := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPingMongo()
	if err = mongoClient.Ping(ctxPingMongo, nil); err != nil {
		appLogger.Fatal("Failed to ping MongoDB", zap.Error(err))
	}
	appLogger.Info("Successfully connected and pinged MongoDB.")
	db := mongoClient.Database(cfg.MongoDatabase) // Use database name from config

	// 5. Initialize NATS Publisher
	natsPublisher, err := natsAdapter.NewPublisher(cfg.NATSURL, appLogger, serviceName)
	if err != nil {
		appLogger.Fatal("Failed to initialize NATS publisher", zap.Error(err))
	}
	defer natsPublisher.Close()
	appLogger.Info("NATS Publisher initialized.")

	// 6. Initialize Repositories
	reviewRepo, err := mongoRepo.NewReviewRepository(db, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize ReviewRepository", zap.Error(err))
	}
	appLogger.Info("ReviewRepository initialized.")

	// 7. Initialize Usecases
	reviewUsecase := usecase.NewReviewUsecase(reviewRepo, natsPublisher, appLogger) // Pass NATS publisher
	appLogger.Info("ReviewUsecase initialized.")

	// 8. Initialize gRPC Handler
	reviewGRPCHandler := grpcAdapter.NewReviewHandler(reviewUsecase, appLogger)
	appLogger.Info("gRPC ReviewHandler initialized.")

	// 9. Start gRPC Server
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		appLogger.Fatal("Failed to listen for gRPC", zap.String("port", cfg.GRPCPort), zap.Error(err))
	}

	// Create gRPC server with interceptors
	grpcSrv := grpcAdapter.NewGRPCServer(appLogger, cfg.JWTSecret, tp) // This now returns *grpc.Server
	pb.RegisterReviewServiceServer(grpcSrv, reviewGRPCHandler)

	go func() {
		appLogger.Info("Starting gRPC server", zap.String("port", cfg.GRPCPort))
		if err := grpcSrv.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			appLogger.Fatal("gRPC server Serve error", zap.Error(err))
		}
	}()

	// 10. Start Prometheus Metrics Server
	if cfg.PrometheusMetricsPort != "" {
		metricsManager := metrics.NewMetricsManager(serviceName) // Initialize metrics
		go func() {
			appLogger.Info("Starting Prometheus metrics server", zap.String("port", cfg.PrometheusMetricsPort))
			if err := metrics.StartMetricsServer(cfg.PrometheusMetricsPort, appLogger, metricsManager.Registry); err != nil && !errors.Is(err, http.ErrServerClosed) {
				appLogger.Error("Prometheus metrics server failed", zap.Error(err))
			}
		}()
		appLogger.Info("Prometheus metrics server configured.")
	} else {
		appLogger.Info("Prometheus metrics server not started (PROMETHEUS_METRICS_PORT not set).")
	}

	// 11. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	appLogger.Info("Received shutdown signal", zap.String("signal", sig.String()))

	appLogger.Info("gRPC health status set to NOT_SERVING")

	// Gracefully stop the gRPC server
	appLogger.Info("Shutting down gRPC server...")
	grpcSrv.GracefulStop()
	appLogger.Info("gRPC server stopped.")
	appLogger.Info("Application shutting down...")
}
