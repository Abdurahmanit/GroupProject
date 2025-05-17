package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/listing/domain"
)

type ListingCache struct {
	client *redis.Client
}

func NewListingCache(addr string) (*ListingCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr, // e.g., "localhost:6379"
	})
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	return &ListingCache{client: client}, nil
}

func (c *ListingCache) GetListing(ctx context.Context, id string) (*domain.Listing, error) {
	data, err := c.client.Get(ctx, "listing:"+id).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, err
	}
	var listing domain.Listing
	if err := json.Unmarshal(data, &listing); err != nil {
		return nil, err
	}
	return &listing, nil
}

func (c *ListingCache) SetListing(ctx context.Context, listing *domain.Listing) error {
	data, err := json.Marshal(listing)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "listing:"+listing.ID, data, 1*time.Hour).Err()
}

func (c *ListingCache) DeleteListing(ctx context.Context, id string) error {
	return c.client.Del(ctx, "listing:"+id).Err()
}