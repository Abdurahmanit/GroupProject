package cache

import (
	"context"
	"time"
)

type CacheRepository interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}
type CacheError string

func (e CacheError) Error() string {
	return string(e)
}

const ErrNotFound = CacheError("key not found in cache")
