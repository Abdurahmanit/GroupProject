package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Abdurahmanit/GroupProject/order-service/internal/domain/entity"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/repository"
	"github.com/redis/go-redis/v9"
)

const (
	cartKeyPrefix = "cart:"
)

type cartRepository struct {
	client *redis.Client
}

func NewCartRepository(client *redis.Client) repository.CartRepository {
	return &cartRepository{
		client: client,
	}
}

func (r *cartRepository) getCartKey(userID string) string {
	return cartKeyPrefix + userID
}

func (r *cartRepository) GetByUserID(ctx context.Context, userID string) (*entity.Cart, error) {
	key := r.getCartKey(userID)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return entity.NewCart(userID), nil
		}
		return nil, fmt.Errorf("failed to get cart for user %s from redis: %w", userID, err)
	}

	var cart entity.Cart
	err = json.Unmarshal([]byte(val), &cart)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cart data for user %s: %w", userID, err)
	}
	return &cart, nil
}

func (r *cartRepository) Save(ctx context.Context, cart *entity.Cart, ttl time.Duration) error {
	if cart == nil || cart.UserID == "" {
		return errors.New("cannot save nil cart or cart with empty userID")
	}

	key := r.getCartKey(cart.UserID)
	cart.UpdatedAt = time.Now().UTC()

	data, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("failed to marshal cart for user %s: %w", cart.UserID, err)
	}

	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to save cart for user %s to redis: %w", cart.UserID, err)
	}
	return nil
}

func (r *cartRepository) DeleteByUserID(ctx context.Context, userID string) error {
	key := r.getCartKey(userID)
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cart for user %s from redis: %w", userID, err)
	}
	return nil
}
