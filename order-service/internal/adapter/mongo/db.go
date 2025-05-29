package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/Abdurahmanit/GroupProject/order-service/internal/app/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	connectTimeout = 10 * time.Second
	pingTimeout    = 5 * time.Second
)

func NewClient(ctx context.Context, cfg config.MongoDBConfig) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(cfg.URI)

	if cfg.User != "" && cfg.Password != "" {
		credential := options.Credential{
			Username: cfg.User,
			Password: cfg.Password,
		}
		clientOptions.SetAuth(credential)
	}

	connectCtx, cancelConnect := context.WithTimeout(ctx, connectTimeout)
	defer cancelConnect()

	client, err := mongo.Connect(connectCtx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	pingCtx, cancelPing := context.WithTimeout(ctx, pingTimeout)
	defer cancelPing()

	err = client.Ping(pingCtx, readpref.Primary())
	if err != nil {
		if disconnectErr := client.Disconnect(context.Background()); disconnectErr != nil {
			// Можно залогировать ошибку дисконнекта, но основная ошибка - пинг
		}
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	return client, nil
}
