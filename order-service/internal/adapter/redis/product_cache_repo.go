package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	listingpb "github.com/Abdurahmanit/GroupProject/listing-service/genproto/listing_service"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/repository"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	productDetailCacheKeyPrefix = "product_detail:"
)

type productDetailCacheRepository struct {
	client *redis.Client
}

func NewProductDetailCacheRepository(client *redis.Client) repository.ProductDetailCache {
	return &productDetailCacheRepository{
		client: client,
	}
}

func (r *productDetailCacheRepository) getProductDetailKey(productID string) string {
	return productDetailCacheKeyPrefix + productID
}

func (r *productDetailCacheRepository) Get(ctx context.Context, productID string) (*listingpb.ListingResponse, error) {
	key := r.getProductDetailKey(productID)
	val, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get product detail for productID %s from redis: %w", productID, err)
	}

	var productDetails listingpb.ListingResponse
	err = proto.Unmarshal(val, &productDetails)
	if err != nil {
		_ = r.Delete(ctx, productID)
		return nil, fmt.Errorf("failed to unmarshal product detail data for productID %s: %w", productID, err)
	}
	return &productDetails, nil
}

func (r *productDetailCacheRepository) Set(ctx context.Context, productID string, productDetails *listingpb.ListingResponse, ttl time.Duration) error {
	if productDetails == nil || productID == "" {
		return errors.New("cannot cache nil product details or product details with empty productID")
	}
	key := r.getProductDetailKey(productID)

	data, err := proto.Marshal(productDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal product details for productID %s: %w", productID, err)
	}

	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set product detail for productID %s to redis: %w", productID, err)
	}
	return nil
}

func (r *productDetailCacheRepository) Delete(ctx context.Context, productID string) error {
	key := r.getProductDetailKey(productID)
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete product detail for productID %s from redis: %w", productID, err)
	}
	return nil
}

func (r *productDetailCacheRepository) getUsingJSON(ctx context.Context, productID string) (*listingpb.ListingResponse, error) {
	key := r.getProductDetailKey(productID)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get product detail (JSON) for productID %s from redis: %w", productID, err)
	}

	var productDetails listingpb.ListingResponse
	unmarshalOpts := protojson.UnmarshalOptions{DiscardUnknown: true}
	err = unmarshalOpts.Unmarshal([]byte(val), &productDetails)
	if err != nil {
		_ = r.Delete(ctx, productID)
		return nil, fmt.Errorf("failed to unmarshal product detail data (JSON) for productID %s: %w", productID, err)
	}
	return &productDetails, nil
}

func (r *productDetailCacheRepository) setUsingJSON(ctx context.Context, productID string, productDetails *listingpb.ListingResponse, ttl time.Duration) error {
	if productDetails == nil || productID == "" {
		return errors.New("cannot cache nil product details or product details with empty productID (JSON)")
	}
	key := r.getProductDetailKey(productID)

	marshalOpts := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: false}
	data, err := marshalOpts.Marshal(productDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal product details (JSON) for productID %s: %w", productID, err)
	}

	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set product detail (JSON) for productID %s to redis: %w", productID, err)
	}
	return nil
}
