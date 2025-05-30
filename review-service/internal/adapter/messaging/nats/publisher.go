package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Abdurahmanit/GroupProject/review-service/internal/platform/logger"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("review-service/nats-publisher")

type Publisher struct {
	conn   *nats.Conn
	logger *logger.Logger
}

func NewPublisher(url string, log *logger.Logger, appName string) (*Publisher, error) {
	log.Info("NATS Publisher: connecting...", zap.String("url", url))

	opts := []nats.Option{
		nats.Name(fmt.Sprintf("%s NATS Publisher", appName)),
		nats.Timeout(10 * time.Second), // Example timeout
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			log.Error("NATS error", zap.Stringp("subject", &sub.Subject), zap.Error(err))
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

type NATSHeaderCarrier nats.Header

func (c NATSHeaderCarrier) Get(key string) string {
	return nats.Header(c).Get(key)
}

func (c NATSHeaderCarrier) Set(key string, value string) {
	nats.Header(c).Set(key, value)
}

func (c NATSHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
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
