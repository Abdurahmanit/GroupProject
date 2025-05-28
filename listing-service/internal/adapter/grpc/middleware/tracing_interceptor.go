// internal/adapter/grpc/middleware/tracing_interceptor.go
package middleware

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

// TracingInterceptor возвращает gRPC унарный interceptor для OpenTelemetry трейсинга.
func TracingInterceptor() grpc.UnaryServerInterceptor {
	// Этот вызов использует глобально сконфигурированные TracerProvider и Propagator.
	// Убедись, что otel.SetTracerProvider() и otel.SetTextMapPropagator()
	// были вызваны ранее (например, в main.go при инициализации трейсера).
	return otelgrpc.UnaryServerInterceptor()
}

// Если тебе нужны Stream Interceptors:
// func StreamTracingInterceptor() grpc.StreamServerInterceptor {
// 	return otelgrpc.StreamServerInterceptor()
// }