package nats

import (
	"fmt"
	"time"

	"github.com/Abdurahmanit/GroupProject/order-service/internal/app/config"
	"github.com/nats-io/nats.go"
)

const (
	connectWait   = 5 * time.Second
	maxReconnects = 5
	reconnectWait = 2 * time.Second
)

func NewConnection(cfg config.NATSConfig) (*nats.Conn, error) {
	opts := []nats.Option{
		nats.Name("OrderService NATS Publisher"),
		nats.Timeout(connectWait),
		nats.MaxReconnects(maxReconnects),
		nats.ReconnectWait(reconnectWait),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
		}),
	}

	nc, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS at %s: %w", cfg.URL, err)
	}

	return nc, nil
}
