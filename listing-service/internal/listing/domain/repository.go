package domain

import "context"

type ListingRepository interface {
	Create(ctx context.Context, listing *Listing) error
	Update(ctx context.Context, listing *Listing) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*Listing, error)
	FindByFilter(ctx context.Context, filter Filter) ([]*Listing, error)
}

type FavoriteRepository interface {
	Add(ctx context.Context, favorite *Favorite) error
	Remove(ctx context.Context, userID, listingID string) error
	FindByUserID(ctx context.Context, userID string) ([]*Favorite, error)
}

type Filter struct {
	Query    string
	MinPrice float64
	MaxPrice float64
	Status   ListingStatus
}