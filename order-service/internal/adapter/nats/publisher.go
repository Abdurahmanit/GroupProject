package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

type MessagePublisher interface {
	Publish(ctx context.Context, subject string, message interface{}) error
	PublishRaw(ctx context.Context, subject string, data []byte) error
}

type natsPublisher struct {
	conn *nats.Conn
}

func NewNATSPublisher(conn *nats.Conn) (MessagePublisher, error) {
	if conn == nil {
		return nil, fmt.Errorf("NATS connection cannot be nil")
	}
	return &natsPublisher{
		conn: conn,
	}, nil
}

func (p *natsPublisher) Publish(ctx context.Context, subject string, message interface{}) error {
	if p.conn == nil {
		return fmt.Errorf("NATS connection is not initialized")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message to JSON for subject %s: %w", subject, err)
	}

	return p.PublishRaw(ctx, subject, data)
}

func (p *natsPublisher) PublishRaw(ctx context.Context, subject string, data []byte) error {
	if p.conn == nil {
		return fmt.Errorf("NATS connection is not initialized")
	}

	if err := p.conn.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish message to NATS subject %s: %w", subject, err)
	}

	return nil
}
