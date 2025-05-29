package grpc

import (
	"github.com/Abdurahmanit/GroupProject/review-service/internal/middleware"
	"github.com/Abdurahmanit/GroupProject/review-service/internal/platform/logger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// NewGRPCServer creates a new gRPC server instance with standard interceptors.
// This function was used by cmd/main.go in listing-service.
// For more granular control over interceptors based on test needs,
// NewGRPCServerWithInterceptors is introduced below.
func NewGRPCServer(
	appLogger *logger.Logger,
	jwtSecret string,
	tp *sdktrace.TracerProvider, // Optional tracer provider
) (*grpc.Server, func()) {
	// Define default public methods and required roles if this server needs its own defaults.
	// However, it's often better to pass these from cmd/main.go for clarity.
	publicMethods := map[string]bool{
		"/review.ReviewService/GetReview":               true,
		"/review.ReviewService/ListReviewsByProduct":    true,
		"/review.ReviewService/GetProductAverageRating": true,
		// Health check is typically public by default or handled by grpc library
		grpc_health_v1.Health_Check_FullMethodName: true,
	}
	requiredRoles := map[string][]string{
		"/review.ReviewService/ModerateReview": {"admin"},
		// Add other admin-specific methods here
	}

	return NewGRPCServerWithInterceptors(appLogger, jwtSecret, tp, publicMethods, requiredRoles)
}

// NewGRPCServerWithInterceptors creates a new gRPC server instance,
// allowing explicit configuration of interceptors, public methods, and required roles.
// This is more flexible for both production and testing.
func NewGRPCServerWithInterceptors(
	appLogger *logger.Logger,
	jwtSecret string,
	tp *sdktrace.TracerProvider, // Optional tracer provider, can be nil
	publicMethods map[string]bool,
	requiredRoles map[string][]string,
) *grpc.Server { // Removed func() from return as cleanup is handled in main

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		middleware.TracingInterceptor(), // Assumes global OTel setup if tp is nil here, or uses tp if provided
		middleware.LoggingInterceptor(appLogger),
		middleware.AuthInterceptor(jwtSecret, appLogger, publicMethods, requiredRoles),
	}

	streamInterceptors := []grpc.StreamServerInterceptor{
		middleware.StreamTracingInterceptor(), // Add stream tracing if you have streaming RPCs
		// Add other stream interceptors if needed
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	)

	appLogger.Info("gRPC server configured with interceptors: Tracing, Logging, Auth")

	// Register reflection service (useful for tools like grpcurl)
	reflection.Register(server)

	// Register gRPC Health Checking Protocol service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	// Initial status can be set in cmd/main.go after server starts listening
	// healthServer.SetServingStatus("review-service", grpc_health_v1.HealthCheckResponse_SERVING)

	return server
	// Cleanup (like tracerProvider.Shutdown) should be handled by the caller (cmd/main.go)
	// to ensure it happens at the very end of the application lifecycle.
}
