package main

import (
	"fmt"
	"net/http"

	"github.com/Abdurahmanit/GroupProject/api-gateway/internal/config"
	"github.com/Abdurahmanit/GroupProject/api-gateway/internal/handler"
	"github.com/Abdurahmanit/GroupProject/api-gateway/internal/middleware"
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

	// Set up router
	r := chi.NewRouter()
	r.Use(middleware.Logger(logger))

	// Public routes
	r.Post("/api/user/register", userHandler.Register)
	r.Post("/api/user/login", userHandler.Login)

	// Protected routes (require JWT authentication)
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuth(cfg.JWTSecret))
		r.Post("/api/user/logout", userHandler.Logout)
		r.Get("/api/user/profile", userHandler.GetProfile)
		r.Put("/api/user/profile", userHandler.UpdateProfile)
		r.Post("/api/user/change-password", userHandler.ChangePassword)
		r.Post("/api/user/verify-email", userHandler.VerifyEmail)
		r.Delete("/api/user/delete", userHandler.DeleteUser)

		// Admin routes
		r.Post("/api/admin/user/delete", userHandler.AdminDeleteUser)
		r.Post("/api/admin/users/list", userHandler.AdminListUsers)
		r.Post("/api/admin/users/search", userHandler.AdminSearchUsers)
		r.Post("/api/admin/user/update-role", userHandler.AdminUpdateUserRole)
	})

	// Start HTTP server
	addr := fmt.Sprintf(":%d", cfg.Port)
	logger.Info("Starting API Gateway", zap.String("address", addr))
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
