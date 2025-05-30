package domain

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReviewRepository interface {
	Create(ctx context.Context, review *Review) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*Review, error)
	Update(ctx context.Context, review *Review) error
	Delete(ctx context.Context, id primitive.ObjectID) error

	FindByProductID(ctx context.Context, productID string, filter ReviewFilter) ([]*Review, int64, error)

	FindByUserID(ctx context.Context, userID string, filter ReviewFilter) ([]*Review, int64, error)

	GetAverageRating(ctx context.Context, productID string) (average float64, count int32, err error)

	FindByStatus(ctx context.Context, status ReviewStatus, filter ReviewFilter) ([]*Review, int64, error)
}
