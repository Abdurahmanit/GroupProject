package tracer

import (
	"context"
	"time"

	"github.com/Abdurahmanit/GroupProject/review-service/internal/platform/logger" // Adjust path
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0" // Use a recent semantic conventions version
	zap "go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // For OTLP gRPC exporter without TLS
)

// InitTracer initializes an OpenTelemetry tracer provider and sets it as the global provider.
// serviceName: The name of the current service (e.g., "review-service").
// otlpEndpoint: The OTLP collector endpoint (e.g., "otel-collector:4317").
// logger: An instance of your application logger.
// Returns the tracer provider for graceful shutdown.
func InitTracer(serviceName, otlpEndpoint string, appLogger *logger.Logger) *sdktrace.TracerProvider {
	if otlpEndpoint == "" {
		appLogger.Info("OpenTelemetry tracing is disabled: OTEL_EXPORTER_OTLP_ENDPOINT is not set.")
		// Return a no-op tracer provider if tracing is disabled
		return sdktrace.NewTracerProvider()
	}

	appLogger.Info("Initializing OpenTelemetry Tracer...",
		zap.String("service_name", serviceName),
		zap.String("otlp_endpoint", otlpEndpoint),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Timeout for exporter setup
	defer cancel()

	// Configure OTLP gRPC exporter
	// Ensure the OTLP collector is accessible at otlpEndpoint
	conn, err := grpc.DialContext(ctx, otlpEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // Use insecure for local/dev if appropriate
		grpc.WithBlock(), // Block until connection is up or context times out
	)
	if err != nil {
		appLogger.Error("Failed to connect to OTLP gRPC collector", zap.Error(err), zap.String("endpoint", otlpEndpoint))
		return sdktrace.NewTracerProvider() // Return no-op provider on failure
	}

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		appLogger.Error("Failed to create OTLP trace exporter", zap.Error(err))
		conn.Close()
		return sdktrace.NewTracerProvider()
	}

	// Define the service resource attributes
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			// Add other relevant resource attributes, e.g., service version, environment
			// semconv.ServiceVersionKey.String("1.0.0"),
			// semconv.DeploymentEnvironmentKey.String("production"),
		),
	)
	if err != nil {
		appLogger.Error("Failed to create OpenTelemetry resource", zap.Error(err))
		traceExporter.Shutdown(ctx) // Clean up exporter
		conn.Close()
		return sdktrace.NewTracerProvider()
	}

	// Create a new tracer provider with the batch span processor and the OTLP exporter.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		// Consider adding a sampler for production environments, e.g., TraceIDRatioBased
		// sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.1)), // Sample 10% of traces
	)

	// Set the global tracer provider.
	otel.SetTracerProvider(tp)

	// Set the global text map propagator (used for context propagation, e.g., HTTP headers, gRPC metadata).
	// W3C Trace Context and W3C Baggage are common choices.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	appLogger.Info("OpenTelemetry Tracer initialized and set as global provider.", zap.String("service_name", serviceName))
	return tp
}
