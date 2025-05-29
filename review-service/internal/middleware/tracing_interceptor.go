package middleware

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	otel "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
)

// TracingInterceptor returns a gRPC unary server interceptor for OpenTelemetry tracing.
// It assumes that the global TracerProvider and Propagator have been set up
// (e.g., in cmd/main.go via tracer.InitTracer).
func TracingInterceptor() grpc.UnaryServerInterceptor {
	return otelgrpc.UnaryServerInterceptor(
		otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
		otelgrpc.WithPropagators(otel.GetTextMapPropagator()),
	)
}

// StreamTracingInterceptor returns a gRPC stream server interceptor for OpenTelemetry tracing.
func StreamTracingInterceptor() grpc.StreamServerInterceptor {
	return otelgrpc.StreamServerInterceptor(
		otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
		otelgrpc.WithPropagators(otel.GetTextMapPropagator()),
	)
}
