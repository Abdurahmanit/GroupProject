package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Abdurahmanit/GroupProject/api-gateway/internal/config"
	"github.com/Abdurahmanit/GroupProject/api-gateway/internal/handler"
	"github.com/Abdurahmanit/GroupProject/api-gateway/internal/middleware"
	"github.com/Abdurahmanit/GroupProject/api-gateway/internal/router"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Инициализация логгера
	logger, _ := zap.NewProduction()
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Error syncing logger: %v\n", err)
		}
	}()

	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load API Gateway config", zap.Error(err))
	}

	userConnAddr := fmt.Sprintf("%s:%d", cfg.UserServiceHost, cfg.UserServicePort)
	userConn, err := grpc.NewClient(userConnAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to connect to User Service", zap.String("address", userConnAddr), zap.Error(err))
	}
	defer userConn.Close()
	logger.Info("Successfully connected to User Service", zap.String("address", userConnAddr))

	// Подключение к Listing Service
	listingConnAddr := fmt.Sprintf("%s:%d", cfg.ListingServiceHost, cfg.ListingServicePort)
	listingConn, err := grpc.NewClient(listingConnAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to connect to Listing Service", zap.String("address", listingConnAddr), zap.Error(err))
	}
	defer listingConn.Close()
	logger.Info("Successfully connected to Listing Service", zap.String("address", listingConnAddr))

	// Подключение к Review Service (Новое)
	reviewConnAddr := fmt.Sprintf("%s:%d", cfg.ReviewServiceHost, cfg.ReviewServicePort)
	reviewConn, err := grpc.NewClient(reviewConnAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to connect to Review Service", zap.String("address", reviewConnAddr), zap.Error(err))
	}
	defer reviewConn.Close()
	logger.Info("Successfully connected to Review Service", zap.String("address", reviewConnAddr))

	// Инициализация обработчиков (сохраняем существующий стиль)
	userHandler := handler.NewUserHandler(userConn, logger)
	listingHandler := handler.NewListingHandler(listingConn, logger)
	reviewHandler := handler.NewReviewHandler(reviewConn, logger)

	r := chi.NewRouter()
	r.Use(middleware.Logger(logger))
	router.SetupUserRoutes(r, userHandler, cfg.JWTSecret)
	router.SetupListingRoutes(r, listingHandler, cfg.JWTSecret)
	router.SetupReviewRoutes(r, reviewHandler, cfg.JWTSecret)

	// Запуск HTTP сервера
	httpServerAddr := fmt.Sprintf(":%d", cfg.Port)
	logger.Info("Starting API Gateway HTTP server", zap.String("address", httpServerAddr))
	if err := http.ListenAndServe(httpServerAddr, r); err != nil {
		logger.Fatal("Failed to start API Gateway HTTP server", zap.Error(err))
	}
}
