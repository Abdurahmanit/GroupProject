package grpc

import (
	"log"

	"google.golang.org/grpc"
	"github.com/your-org/bike-store/listing-service/internal/adapter/grpc/middleware"
	"github.com/your-org/bike-store/listing-service/internal/platform/logger"
	"github.com/your-org/bike-store/listing-service/internal/platform/tracer"
)

func NewGRPCServer() *grpc.Server {
	// Initialize logger and tracer
	logger := logger.NewLogger()
	tracerProvider := tracer.InitTracer()

	// Create gRPC server with middleware
	server := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.LoggingInterceptor(logger)),
	)

	// Ensure tracer shutdown on server stop
	go func() {
		<-server.GetServiceInfo()
		tracerProvider.Shutdown()
	}()

	log.Println("gRPC server initialized with logging middleware")
	return server
}