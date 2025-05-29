package middleware

// Definitions for UserIDKeyType, UserRoleKeyType, UserIDKey, UserRoleKey, Claims, and AuthInterceptor
// have been REMOVED from this file as they are defined in auth_interceptor.go
// and should not be duplicated within the same package.

import (
	"context"
	"time"

	"github.com/Abdurahmanit/GroupProject/review-service/internal/platform/logger" // Adjust path if necessary
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap" // Import zap
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor creates a gRPC unary server interceptor for logging requests.
func LoggingInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		startTime := time.Now()

		span := trace.SpanFromContext(ctx)
		traceID := span.SpanContext().TraceID().String()
		spanID := span.SpanContext().SpanID().String()

		// Extract UserID and Role from context if set by AuthInterceptor
		userIDVal := ctx.Value(UserIDKey) // UserIDKey is accessible as it's in the same package
		roleVal := ctx.Value(UserRoleKey) // UserRoleKey is accessible

		var userIDStr, roleStr string
		if userIDVal != nil {
			userIDStr, _ = userIDVal.(string)
		}
		if roleVal != nil {
			roleStr, _ = roleVal.(string)
		}

		initialLogFields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Time("start_time_utc", startTime.UTC()),
			zap.String("trace_id", traceID),
			zap.String("span_id", spanID),
		}
		if userIDStr != "" {
			initialLogFields = append(initialLogFields, zap.String("user_id", userIDStr))
		}
		if roleStr != "" {
			initialLogFields = append(initialLogFields, zap.String("user_role", roleStr))
		}

		log.Info("gRPC request received", initialLogFields...)

		resp, err := handler(ctx, req)

		duration := time.Since(startTime)
		statusCode := status.Code(err)

		finalLogFields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Duration("duration_ms", duration),
			zap.String("status_code", statusCode.String()),
			zap.String("trace_id", traceID),
			zap.String("span_id", spanID),
		}
		if userIDStr != "" {
			finalLogFields = append(finalLogFields, zap.String("user_id", userIDStr))
		}
		if roleStr != "" {
			finalLogFields = append(finalLogFields, zap.String("user_role", roleStr))
		}

		if err != nil {
			finalLogFields = append(finalLogFields, zap.Error(err))
			log.Error("gRPC request failed", finalLogFields...)
		} else {
			log.Info("gRPC request completed", finalLogFields...)
		}

		return resp, err
	}
}
