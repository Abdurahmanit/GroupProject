package middleware

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel" // Correct import for otel global functions
	"google.golang.org/grpc"
)

// TracingInterceptor returns a gRPC unary server interceptor for OpenTelemetry tracing.
func TracingInterceptor() grpc.UnaryServerInterceptor {
	return otelgrpc.UnaryServerInterceptor(
		otelgrpc.WithTracerProvider(otel.GetTracerProvider()), // Corrected: Use otel.GetTracerProvider()
		otelgrpc.WithPropagators(otel.GetTextMapPropagator()), // Corrected: Use otel.GetTextMapPropagator()
	)
}

// StreamTracingInterceptor returns a gRPC stream server interceptor for OpenTelemetry tracing.
func StreamTracingInterceptor() grpc.StreamServerInterceptor {
	return otelgrpc.StreamServerInterceptor(
		otelgrpc.WithTracerProvider(otel.GetTracerProvider()), // Corrected
		otelgrpc.WithPropagators(otel.GetTextMapPropagator()), // Corrected
	)
}
