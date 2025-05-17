package main

import (
	"fmt"
	"net/http"

	"github.com/Abdurahmanit/GroupProject/api-gateway/internal/config"
	"github.com/Abdurahmanit/GroupProject/api-gateway/internal/handler"
	"github.com/Abdurahmanit/GroupProject/api-gateway/internal/middleware"
	"github.com/Abdurahmanit/GroupProject/api-gateway/internal/router" // Import the router package
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Connect to User Service via gRPC
	userConn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", cfg.UserServiceHost, cfg.UserServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Fatal("Failed to connect to User Service", zap.Error(err))
	}
	defer userConn.Close()

	// Initialize handlers
	userHandler := handler.NewUserHandler(userConn, logger)
	// Add other service handlers here

	// Set up main router
	r := chi.NewRouter()
	r.Use(middleware.Logger(logger))

	router.SetupUserRoutes(r, userHandler, cfg.JWTSecret)
	// Setup other routes

	// Start HTTP server
	addr := fmt.Sprintf(":%d", cfg.Port)
	logger.Info("Starting API Gateway", zap.String("address", addr))
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
