package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time" // Added for nats.Timeout

	"github.com/Abdurahmanit/GroupProject/review-service/internal/platform/logger" // Adjust path if necessary
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel" // Required for NATSHeaderCarrier
	"go.uber.org/zap"          // Import zap
)

var tracer = otel.Tracer("review-service/nats-publisher")

// Publisher handles publishing messages to NATS.
type Publisher struct {
	conn   *nats.Conn
	logger *logger.Logger
}

// NewPublisher creates a new NATS publisher.
func NewPublisher(url string, log *logger.Logger, appName string) (*Publisher, error) {
	log.Info("NATS Publisher: connecting...", zap.String("url", url))

	opts := []nats.Option{
		nats.Name(fmt.Sprintf("%s NATS Publisher", appName)),
		nats.Timeout(10 * time.Second), // Example timeout
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			log.Error("NATS error", zap.Stringp("subject", &sub.Subject), zap.Error(err)) // Use zap.Stringp for potentially nil subject
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Info("NATS connection closed")
		}),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) { // Corrected
			log.Warn("NATS disconnected", zap.Error(err))
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) { // Corrected
			log.Info("NATS reconnected", zap.String("url", nc.ConnectedUrl()))
		}),
	}

	conn, err := nats.Connect(url, opts...)
	if err != nil {
		log.Error("NATS Publisher: failed to connect", zap.String("url", url), zap.Error(err))
		return nil, fmt.Errorf("failed to connect to NATS at %s: %w", url, err)
	}
	log.Info("NATS Publisher: successfully connected", zap.String("url", conn.ConnectedUrl()))

	return &Publisher{
		conn:   conn,
		logger: log.Named("NATSPublisher"),
	}, nil
}

// Publish sends a message to the specified NATS subject.
// It injects OpenTelemetry trace context into the message headers.
func (p *Publisher) Publish(ctx context.Context, subject string, data interface{}) error {
	_, span := tracer.Start(ctx, fmt.Sprintf("NATS.Publish.%s", subject))
	defer span.End()

	p.logger.Debug("NATS Publisher: publishing message", zap.String("subject", subject))

	jsonData, err := json.Marshal(data)
	if err != nil {
		p.logger.Error("NATS Publisher: failed to marshal data to JSON", zap.String("subject", subject), zap.Error(err))
		span.RecordError(err)
		return fmt.Errorf("failed to marshal data for subject %s: %w", subject, err)
	}

	msg := nats.NewMsg(subject)
	msg.Data = jsonData
	msg.Header = make(nats.Header) // nats.Header is map[string][]string

	propagator := otel.GetTextMapPropagator()
	// Correct way to inject into nats.Header (which is http.Header alias)
	propagator.Inject(ctx, NATSHeaderCarrier(msg.Header))

	err = p.conn.PublishMsg(msg)
	if err != nil {
		p.logger.Error("NATS Publisher: failed to publish message", zap.String("subject", subject), zap.Error(err))
		span.RecordError(err)
		return fmt.Errorf("failed to publish message to subject %s: %w", subject, err)
	}

	p.logger.Info("NATS Publisher: message published successfully", zap.String("subject", subject), zap.Int("data_size_bytes", len(jsonData)))
	return nil
}

// NATSHeaderCarrier adapts nats.Header (which is an alias for http.Header) to be a TextMapCarrier.
type NATSHeaderCarrier nats.Header // Note: nats.Header is an alias for http.Header

// Get returns the value associated with the passed key.
// It is used by the OpenTelemetry propagator to extract context.
func (c NATSHeaderCarrier) Get(key string) string {
	return nats.Header(c).Get(key)
}

// Set stores the key-value pair.
// It is used by the OpenTelemetry propagator to inject context.
func (c NATSHeaderCarrier) Set(key string, value string) {
	nats.Header(c).Set(key, value)
}

// Keys returns a slice of all keys in the carrier.
// It is used by the OpenTelemetry propagator.
func (c NATSHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	// http.Header is map[string][]string. The propagator expects keys to be unique.
	// The default TextMapPropagator (TraceContext and Baggage) uses specific keys.
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

// Close drains and closes the NATS connection.
func (p *Publisher) Close() {
	p.logger.Info("NATS Publisher: closing connection...")
	if p.conn != nil && !p.conn.IsClosed() {
		if err := p.conn.Drain(); err != nil {
			p.logger.Error("NATS Publisher: failed to drain connection", zap.Error(err))
		}
		p.conn.Close()
		p.logger.Info("NATS Publisher: connection closed.")
	} else {
		p.logger.Info("NATS Publisher: connection already closed or not initialized.")
	}
}
