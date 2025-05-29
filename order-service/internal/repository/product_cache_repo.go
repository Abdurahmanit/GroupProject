package repository

import (
	"context"
	"time"

	listingpb "github.com/Abdurahmanit/GroupProject/listing-service/genproto/listing_service"
)

type ProductDetailCache interface {
	Get(ctx context.Context, productID string) (*listingpb.ListingResponse, error)
	Set(ctx context.Context, productID string, productDetails *listingpb.ListingResponse, ttl time.Duration) error
	Delete(ctx context.Context, productID string) error
}
