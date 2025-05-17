package grpc

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/adapter/grpc/middleware"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/platform/logger"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/platform/tracer"
)

func NewGRPCServer() (*grpc.Server, func()) {
	// Initialize logger and tracer
	logger := logger.NewLogger()
	tracerProvider := tracer.InitTracer()

	// Create gRPC server with middleware
	server := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.LoggingInterceptor(logger)),
	)

	// Return server and cleanup function
	cleanup := func() {
		// Shutdown tracer provider with context
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			log.Printf("Failed to shutdown tracer: %v", err)
		}
	}

	log.Println("gRPC server initialized with logging middleware")
	return server, cleanup
}