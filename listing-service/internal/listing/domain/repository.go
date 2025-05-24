package domain

import "context"

type ListingRepository interface {
	Create(ctx context.Context, listing *Listing) error
	Update(ctx context.Context, listing *Listing) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*Listing, error)
	FindByFilter(ctx context.Context, filter Filter) (listings []*Listing, total int64, err error)
}

type FavoriteRepository interface {
	Add(ctx context.Context, favorite *Favorite) error
	Remove(ctx context.Context, userID, listingID string) error
	FindByUserID(ctx context.Context, userID string) ([]*Favorite, error)
}

type Storage interface {
    Upload(ctx context.Context, fileName string, data []byte) (string, error)
    // Delete(ctx context.Context, fileKey string) error // Возможно, другие методы
}

