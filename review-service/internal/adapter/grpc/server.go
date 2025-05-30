package grpc

import (
	"github.com/Abdurahmanit/GroupProject/review-service/internal/middleware"
	"github.com/Abdurahmanit/GroupProject/review-service/internal/platform/logger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func NewGRPCServer(
	appLogger *logger.Logger,
	jwtSecret string,
	tp *sdktrace.TracerProvider,
) *grpc.Server {
	publicMethods := map[string]bool{
		"/review.ReviewService/GetReview":               true,
		"/review.ReviewService/ListReviewsByProduct":    true,
		"/review.ReviewService/GetProductAverageRating": true,
		grpc_health_v1.Health_Check_FullMethodName:      true,
	}
	requiredRoles := map[string][]string{
		"/review.ReviewService/ModerateReview": {"admin"},
	}

	return NewGRPCServerWithInterceptors(appLogger, jwtSecret, tp, publicMethods, requiredRoles)
}

func NewGRPCServerWithInterceptors(
	appLogger *logger.Logger,
	jwtSecret string,
	tp *sdktrace.TracerProvider,
	publicMethods map[string]bool,
	requiredRoles map[string][]string,
) *grpc.Server {

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		middleware.TracingInterceptor(),
		middleware.LoggingInterceptor(appLogger),
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
		zap.Bool("tracing_enabled", tp != nil || middleware.TracingInterceptor() != nil),
		zap.Bool("logging_enabled", true),
		zap.Bool("auth_enabled", true),
	)

	reflection.Register(server)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	return server
}
