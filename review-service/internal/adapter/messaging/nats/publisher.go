package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Abdurahmanit/GroupProject/review-service/internal/platform/logger" // Adjust path if necessary
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	zap "go.uber.org/zap"
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
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			log.Error("NATS error", zap.String("subject", sub.Subject), zap.Error(err))
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Info("NATS connection closed")
		}),
		nats.DisconnectedErrHandler(func(nc *nats.Conn, err error) {
			log.Warn("NATS disconnected", zap.Error(err))
		}),
		nats.ReconnectedHandler(func(nc *nats.Conn) {
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

	// Create a NATS message and inject trace context into headers
	msg := nats.NewMsg(subject)
	msg.Data = jsonData
	msg.Header = make(nats.Header)

	// Inject OpenTelemetry context into NATS headers
	propagator := otel.GetTextMapPropagator()
	carrier := propagation.HeaderCarrier(msg.Header) // NATS header implements TextMapCarrier
	propagator.Inject(ctx, carrier)

	err = p.conn.PublishMsg(msg) // Use PublishMsg to send message with headers
	if err != nil {
		p.logger.Error("NATS Publisher: failed to publish message", zap.String("subject", subject), zap.Error(err))
		span.RecordError(err)
		return fmt.Errorf("failed to publish message to subject %s: %w", subject, err)
	}

	p.logger.Info("NATS Publisher: message published successfully", zap.String("subject", subject), zap.Int("data_size_bytes", len(jsonData)))
	return nil
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
