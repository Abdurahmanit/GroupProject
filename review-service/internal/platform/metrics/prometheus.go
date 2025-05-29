package metrics

import (
	"net/http"

	"github.com/Abdurahmanit/GroupProject/review-service/internal/platform/logger" // Adjust path
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// MetricsManager holds custom Prometheus metrics.
type MetricsManager struct {
	Registry             *prometheus.Registry
	ReviewsCreatedTotal  prometheus.Counter
	ReviewUpdatesTotal   prometheus.Counter
	ReviewDeletesTotal   prometheus.Counter
	ReviewAPIErrorsTotal *prometheus.CounterVec   // To count errors by RPC method
	ReviewAPILatency     *prometheus.HistogramVec // To measure RPC latency by method
	// Add more metrics as needed, e.g., average ratings, moderation actions
}

// NewMetricsManager initializes and registers custom Prometheus metrics.
func NewMetricsManager(serviceName string) *MetricsManager {
	registry := prometheus.NewRegistry() // Use a custom registry to avoid conflicts if multiple services run in one process (rare for microservices)
	// For standard global registry: registry = prometheus.DefaultRegisterer

	reviewsCreatedTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: serviceName,
		Name:      "reviews_created_total",
		Help:      "Total number of reviews created.",
	})
	reviewUpdatesTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: serviceName,
		Name:      "review_updates_total",
		Help:      "Total number of reviews updated.",
	})
	reviewDeletesTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: serviceName,
		Name:      "review_deletes_total",
		Help:      "Total number of reviews deleted.",
	})
	reviewAPIErrorsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: serviceName,
		Name:      "api_errors_total",
		Help:      "Total number of API errors by method.",
	}, []string{"method", "error_type"}) // Labels for method and type of error

	reviewAPILatency := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: serviceName,
		Name:      "api_request_latency_seconds",
		Help:      "Latency of API requests by method.",
		Buckets:   prometheus.DefBuckets, // Default buckets: .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
	}, []string{"method"})

	registry.MustRegister(
		reviewsCreatedTotal,
		reviewUpdatesTotal,
		reviewDeletesTotal,
		reviewAPIErrorsTotal,
		reviewAPILatency,
		prometheus.NewGoCollector(), // Standard Go runtime metrics
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}), // Process metrics
	)

	return &MetricsManager{
		Registry:             registry,
		ReviewsCreatedTotal:  reviewsCreatedTotal,
		ReviewUpdatesTotal:   reviewUpdatesTotal,
		ReviewDeletesTotal:   reviewDeletesTotal,
		ReviewAPIErrorsTotal: reviewAPIErrorsTotal,
		ReviewAPILatency:     reviewAPILatency,
	}
}

// StartMetricsServer starts an HTTP server to expose Prometheus metrics.
// port: The port number for the metrics server (e.g., "9093").
// logger: Application logger.
// registry: The Prometheus registry to use for the handler.
func StartMetricsServer(port string, appLogger *logger.Logger, registry *prometheus.Registry) error {
	if port == "" {
		appLogger.Info("Prometheus metrics server port not configured, server will not start.")
		return nil
	}

	mux := http.NewServeMux()
	// Use promhttp.HandlerFor to specify the registry.
	// If using prometheus.DefaultRegisterer, promhttp.Handler() is sufficient.
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	appLogger.Info("Prometheus metrics server starting", zap.String("port", port), zap.String("path", "/metrics"))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	return server.ListenAndServe()
}
