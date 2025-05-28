package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Abdurahmanit/GroupProject/news-service/internal/config"
	"github.com/Abdurahmanit/GroupProject/news-service/internal/entity"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

const (
	NewsCreatedSubject = "news.created"
	NewsUpdatedSubject = "news.updated"
	NewsDeletedSubject = "news.deleted"
)

type Publisher struct {
	nc     *nats.Conn
	logger *zap.Logger
}

type DeletedEventPayload struct {
	ID string `json:"id"`
}

func NewNATSPublisher(cfg *config.NATSConfig, logger *zap.Logger) (*Publisher, error) {
	opts := []nats.Option{
		nats.Timeout(cfg.ConnectTimeout),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			logger.Error("NATS error", zap.String("subject", sub.Subject), zap.Error(err))
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			logger.Info("NATS connection closed")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Info("NATS reconnected", zap.String("url", nc.ConnectedUrl()))
		}),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			logger.Warn("NATS disconnected", zap.Error(err))
		}),
	}

	nc, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	logger.Info("Successfully connected to NATS", zap.String("url", nc.ConnectedUrl()))

	return &Publisher{nc: nc, logger: logger}, nil
}

func (p *Publisher) PublishNewsCreated(ctx context.Context, news *entity.News) error {
	data, err := json.Marshal(news)
	if err != nil {
		p.logger.Error("Failed to marshal news for NATS publishing (created event)",
			zap.Error(err),
			zap.String("news_id", news.ID),
			zap.String("subject", NewsCreatedSubject),
		)
		return fmt.Errorf("failed to marshal news for %s: %w", NewsCreatedSubject, err)
	}

	if err := p.nc.Publish(NewsCreatedSubject, data); err != nil {
		p.logger.Error("Failed to publish NATS message",
			zap.String("subject", NewsCreatedSubject),
			zap.Error(err),
			zap.String("news_id", news.ID),
		)
		return fmt.Errorf("failed to publish NATS message for %s: %w", NewsCreatedSubject, err)
	}
	p.logger.Info("Published NATS message",
		zap.String("subject", NewsCreatedSubject),
		zap.String("news_id", news.ID),
	)
	return nil
}

func (p *Publisher) PublishNewsUpdated(ctx context.Context, news *entity.News) error {
	data, err := json.Marshal(news)
	if err != nil {
		p.logger.Error("Failed to marshal news for NATS publishing (updated event)",
			zap.Error(err),
			zap.String("news_id", news.ID),
			zap.String("subject", NewsUpdatedSubject),
		)
		return fmt.Errorf("failed to marshal news for %s: %w", NewsUpdatedSubject, err)
	}

	if err := p.nc.Publish(NewsUpdatedSubject, data); err != nil {
		p.logger.Error("Failed to publish NATS message",
			zap.String("subject", NewsUpdatedSubject),
			zap.Error(err),
			zap.String("news_id", news.ID),
		)
		return fmt.Errorf("failed to publish NATS message for %s: %w", NewsUpdatedSubject, err)
	}
	p.logger.Info("Published NATS message",
		zap.String("subject", NewsUpdatedSubject),
		zap.String("news_id", news.ID),
	)
	return nil
}

func (p *Publisher) PublishNewsDeleted(ctx context.Context, newsID string) error {
	payload := DeletedEventPayload{ID: newsID}
	data, err := json.Marshal(payload)
	if err != nil {
		p.logger.Error("Failed to marshal news ID for NATS publishing (deleted event)",
			zap.Error(err),
			zap.String("news_id", newsID),
			zap.String("subject", NewsDeletedSubject),
		)
		return fmt.Errorf("failed to marshal news ID for %s: %w", NewsDeletedSubject, err)
	}

	if err := p.nc.Publish(NewsDeletedSubject, data); err != nil {
		p.logger.Error("Failed to publish NATS message",
			zap.String("subject", NewsDeletedSubject),
			zap.Error(err),
			zap.String("news_id", newsID),
		)
		return fmt.Errorf("failed to publish NATS message for %s: %w", NewsDeletedSubject, err)
	}
	p.logger.Info("Published NATS message",
		zap.String("subject", NewsDeletedSubject),
		zap.String("news_id", newsID),
	)
	return nil
}

func (p *Publisher) Close() {
	if p.nc != nil && !p.nc.IsClosed() {
		if err := p.nc.Drain(); err != nil { // Drain ensures all buffered messages are sent
			p.logger.Error("Error draining NATS connection", zap.Error(err))
		}
		p.nc.Close()
		p.logger.Info("NATS publisher connection closed")
	}
}
