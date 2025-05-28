// internal/adapter/messaging/nats/publisher.go
package nats

import (
	"context"
	"encoding/json"
	"fmt" // Для форматирования ошибок

	// Путь к твоему кастомному логгеру
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/platform/logger"
	"github.com/nats-io/nats.go"
)

type Publisher struct {
	conn   *nats.Conn
	logger *logger.Logger // <--- ДОБАВЛЕНО поле для логгера
}

// NewPublisher теперь принимает логгер
func NewPublisher(url string, log *logger.Logger) (*Publisher, error) { // <--- ДОБАВЛЕН параметр log *logger.Logger
	log.Info("NATS Publisher: connecting...", "url", url)
	conn, err := nats.Connect(url,
		// Опции для NATS соединения, если нужны:
		// nats.Name("Listing Service Publisher"),
		// nats.Timeout(5*time.Second),
		// nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
		// 	log.Error("NATS Error", "subject", sub.Subject, "error", err)
		// }),
		// nats.ClosedHandler(func(nc *nats.Conn) {
		// 	log.Info("NATS connection closed")
		// }),
		// nats.DisconnectedErrHandler(func(nc *nats.Conn, err error) {
		// 	log.Warn("NATS disconnected", "error", err)
		// }),
		// nats.ReconnectedHandler(func(nc *nats.Conn) {
		// 	log.Info("NATS reconnected", "url", nc.ConnectedUrl())
		// }),
	)
	if err != nil {
		log.Error("NATS Publisher: failed to connect", "url", url, "error", err)
		return nil, fmt.Errorf("failed to connect to NATS at %s: %w", url, err)
	}
	log.Info("NATS Publisher: successfully connected", "url", conn.ConnectedUrl()) // Используем conn.ConnectedUrl() для фактического URL

	return &Publisher{
		conn:   conn,
		logger: log, // <--- СОХРАНЯЕМ логгер
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, subject string, data interface{}) error {
	p.logger.Debug("NATS Publisher: publishing message", "subject", subject, "data_type", fmt.Sprintf("%T", data))

	jsonData, err := json.Marshal(data)
	if err != nil {
		p.logger.Error("NATS Publisher: failed to marshal data to JSON", "subject", subject, "error", err)
		return fmt.Errorf("failed to marshal data for subject %s: %w", subject, err)
	}

	// Для трейсинга, если нужно, можно передать контекст (некоторые NATS клиенты или обертки это поддерживают)
	// или создать дочерний спан здесь. Стандартный p.conn.Publish не принимает context.
	// Если используется JetStream, то там есть PublishMsg, который принимает nats.Msg с заголовками,
	// куда можно внедрить контекст трейсинга.
	// Для простого Publish, контекст трейсинга обычно не передается напрямую в эту функцию.

	err = p.conn.Publish(subject, jsonData)
	if err != nil {
		p.logger.Error("NATS Publisher: failed to publish message", "subject", subject, "error", err)
		return fmt.Errorf("failed to publish message to subject %s: %w", subject, err)
	}

	p.logger.Info("NATS Publisher: message published successfully", "subject", subject, "data_size_bytes", len(jsonData))
	return nil
}

func (p *Publisher) Close() {
	p.logger.Info("NATS Publisher: closing connection...")
	if p.conn != nil && !p.conn.IsClosed() {
		p.conn.Drain() // Рекомендуется Drain перед Close для отправки буферизованных сообщений
		p.conn.Close()
		p.logger.Info("NATS Publisher: connection closed.")
	} else {
		p.logger.Info("NATS Publisher: connection already closed or not initialized.")
	}
}