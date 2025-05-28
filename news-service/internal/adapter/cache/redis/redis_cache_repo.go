package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Abdurahmanit/GroupProject/news-service/internal/config"
	"github.com/Abdurahmanit/GroupProject/news-service/internal/port/cache"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type redisCacheRepository struct {
	client *redis.Client
	logger *zap.Logger
}

func NewRedisClient(cfg *config.RedisConfig, logger *zap.Logger) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	statusCmd := rdb.Ping(pingCtx)
	if err := statusCmd.Err(); err != nil {
		logger.Error("Failed to connect to Redis", zap.String("address", cfg.Address), zap.Error(err))
		return nil, fmt.Errorf("failed to ping redis at %s: %w", cfg.Address, err)
	}
	logger.Info("Successfully connected to Redis", zap.String("address", cfg.Address))
	return rdb, nil
}

func NewRedisCacheRepository(client *redis.Client, logger *zap.Logger) cache.CacheRepository {
	return &redisCacheRepository{
		client: client,
		logger: logger,
	}
}

func (r *redisCacheRepository) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, cache.ErrNotFound // Используем нашу ошибку из port/cache
		}
		r.logger.Error("Redis Get operation failed", zap.String("key", key), zap.Error(err))
		return nil, fmt.Errorf("redisCacheRepository.Get for key '%s': %w", key, err)
	}
	return val, nil
}

func (r *redisCacheRepository) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		r.logger.Error("Redis Set operation failed", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("redisCacheRepository.Set for key '%s': %w", key, err)
	}
	r.logger.Debug("Redis Set operation successful", zap.String("key", key), zap.Duration("ttl", ttl))
	return nil
}

func (r *redisCacheRepository) Delete(ctx context.Context, key string) error {
	cmdResult := r.client.Del(ctx, key)
	if err := cmdResult.Err(); err != nil {
		r.logger.Error("Redis Del operation failed", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("redisCacheRepository.Delete for key '%s': %w", key, err)
	}
	r.logger.Debug("Redis Del operation attempted (check logs for count if needed)", zap.String("key", key))
	return nil
}
