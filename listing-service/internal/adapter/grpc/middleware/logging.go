package middleware

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/platform/logger"
)

func LoggingInterceptor(logger *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		logger.Info("gRPC request", "method", info.FullMethod, "start_time", start)

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		if err != nil {
			logger.Error("gRPC request failed", "method", info.FullMethod, "duration", duration, "error", err)
		} else {
			logger.Info("gRPC request completed", "method", info.FullMethod, "duration", duration)
		}

		return resp, err
	}
}