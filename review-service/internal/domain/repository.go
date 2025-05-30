package domain

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReviewRepository defines the interface for review data persistence.
// Methods operate on the clean domain.Review entity, without any
// direct knowledge of database-specific tags or structures.
type ReviewRepository interface {
	Create(ctx context.Context, review *Review) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*Review, error)
	Update(ctx context.Context, review *Review) error // For updating comment, rating by author, or status by moderator
	Delete(ctx context.Context, id primitive.ObjectID) error

	// FindByProductID retrieves reviews for a specific product, with pagination and optional status filter.
	// Returns reviews, total count for pagination, and error.
	FindByProductID(ctx context.Context, productID string, filter ReviewFilter) ([]*Review, int64, error)

	// FindByUserID retrieves reviews written by a specific user, with pagination.
	// Returns reviews, total count for pagination, and error.
	FindByUserID(ctx context.Context, userID string, filter ReviewFilter) ([]*Review, int64, error)

	// GetAverageRating calculates the average rating and count of reviews for a product.
	// Considers only 'Approved' reviews.
	GetAverageRating(ctx context.Context, productID string) (average float64, count int32, err error)

	// FindByStatus retrieves reviews by their status, e.g., for a moderation queue.
	// Returns reviews, total count for pagination, and error.
	FindByStatus(ctx context.Context, status ReviewStatus, filter ReviewFilter) ([]*Review, int64, error)
}
