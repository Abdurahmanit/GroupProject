package grpc

import (
	"github.com/Abdurahmanit/GroupProject/review-service/internal/middleware"
	"github.com/Abdurahmanit/GroupProject/review-service/internal/platform/logger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap" // Import zap for logger fields
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// NewGRPCServer creates a new gRPC server instance with standard interceptors.
// It no longer returns a cleanup function; cleanup is handled by the caller (main.go).
func NewGRPCServer(
	appLogger *logger.Logger,
	jwtSecret string,
	tp *sdktrace.TracerProvider, // Optional tracer provider
) *grpc.Server {
	publicMethods := map[string]bool{
		"/review.ReviewService/GetReview":               true,
		"/review.ReviewService/ListReviewsByProduct":    true,
		"/review.ReviewService/GetProductAverageRating": true,
		grpc_health_v1.Health_Check_FullMethodName:      true, // Health check is public
	}
	requiredRoles := map[string][]string{
		"/review.ReviewService/ModerateReview": {"admin"},
	}

	return NewGRPCServerWithInterceptors(appLogger, jwtSecret, tp, publicMethods, requiredRoles)
}

// NewGRPCServerWithInterceptors creates a new gRPC server instance,
// allowing explicit configuration of interceptors, public methods, and required roles.
func NewGRPCServerWithInterceptors(
	appLogger *logger.Logger,
	jwtSecret string,
	tp *sdktrace.TracerProvider,
	publicMethods map[string]bool,
	requiredRoles map[string][]string,
) *grpc.Server {

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		middleware.TracingInterceptor(),          // Assumes global OTel setup if tp is nil here
		middleware.LoggingInterceptor(appLogger), // Corrected: Use the function from middleware package
		middleware.AuthInterceptor(jwtSecret, appLogger, publicMethods, requiredRoles),
	}

	streamInterceptors := []grpc.StreamServerInterceptor{
		middleware.StreamTracingInterceptor(),
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	)

	appLogger.Info("gRPC server configured with interceptors",
		zap.Bool("tracing_enabled", tp != nil || middleware.TracingInterceptor() != nil), // Check if tracing is active
		zap.Bool("logging_enabled", true),
		zap.Bool("auth_enabled", true),
	)

	reflection.Register(server)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	// Initial serving status is set in main.go after the server starts listening.

	return server
}
