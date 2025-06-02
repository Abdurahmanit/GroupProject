package tracer

import (
	"context"
	"time"

	"github.com/Abdurahmanit/GroupProject/review-service/internal/platform/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	zap "go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitTracer(serviceName, otlpEndpoint string, appLogger *logger.Logger) *sdktrace.TracerProvider {
	if otlpEndpoint == "" {
		appLogger.Info("OpenTelemetry tracing is disabled: OTEL_EXPORTER_OTLP_ENDPOINT is not set.")
		return sdktrace.NewTracerProvider()
	}

	appLogger.Info("Initializing OpenTelemetry Tracer...",
		zap.String("service_name", serviceName),
		zap.String("otlp_endpoint", otlpEndpoint),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Timeout for exporter setup
	defer cancel()

	conn, err := grpc.DialContext(ctx, otlpEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // Use insecure for local/dev if appropriate
		grpc.WithBlock(),
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
		),
	)
	if err != nil {
		appLogger.Error("Failed to create OpenTelemetry resource", zap.Error(err))
		traceExporter.Shutdown(ctx) // Clean up exporter
		conn.Close()
		return sdktrace.NewTracerProvider()
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)

	// Set the global tracer provider.
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	appLogger.Info("OpenTelemetry Tracer initialized and set as global provider.", zap.String("service_name", serviceName))
	return tp
}
