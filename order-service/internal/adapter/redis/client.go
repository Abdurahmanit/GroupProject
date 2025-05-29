package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/Abdurahmanit/GroupProject/order-service/internal/app/config"
	"github.com/redis/go-redis/v9"
)

const (
	dialTimeout = 5 * time.Second
)

func NewClient(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	opts := &redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	client := redis.NewClient(opts)

	dialCtx, cancel := context.WithTimeout(ctx, dialTimeout)
	defer cancel()

	_, err := client.Ping(dialCtx).Result()
	if err != nil {
		// Попытаться закрыть клиент, если пинг не удался
		if closeErr := client.Close(); closeErr != nil {
			// Можно залогировать ошибку закрытия, но основная ошибка - пинг
		}
		return nil, fmt.Errorf("failed to connect to redis (ping failed): %w", err)
	}

	return client, nil
}
