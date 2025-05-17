package nats

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	conn *nats.Conn
}

func NewPublisher(url string) (*Publisher, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	return &Publisher{conn: conn}, nil
}

func (p *Publisher) Publish(ctx context.Context, subject string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return p.conn.Publish(subject, jsonData)
}

func (p *Publisher) Close() {
	p.conn.Close()
}