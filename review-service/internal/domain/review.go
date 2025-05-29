package domain

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive" // Still using ObjectID for ID type in domain
)

// --- Domain Specific Errors ---

var (
	// ErrNotFound indicates that a requested entity was not found.
	ErrNotFound = errors.New("entity not found")
	// ErrForbidden indicates that the user is not authorized to perform the action.
	ErrForbidden = errors.New("action forbidden")
	// ErrInvalidInput indicates that the provided input data is invalid.
	ErrInvalidInput = errors.New("invalid input data")
	// ErrReviewAlreadyExists indicates that a user has already reviewed a specific product/seller.
	ErrReviewAlreadyExists = errors.New("review already exists for this user and target")
	// ErrOptimisticLock indicates a conflict due to concurrent modification.
	ErrOptimisticLock = errors.New("optimistic lock conflict: data was modified by another process")
	// ErrRepository indicates a generic data persistence error.
	ErrRepository = errors.New("repository error")
)

// --- Review Status Enum ---

// ReviewStatus represents the moderation status of a review.
type ReviewStatus string

const (
	ReviewStatusPending  ReviewStatus = "pending"
	ReviewStatusApproved ReviewStatus = "approved"
	ReviewStatusRejected ReviewStatus = "rejected"
	ReviewStatusHidden   ReviewStatus = "hidden"   // Hidden by admin, not deleted
	ReviewStatusReported ReviewStatus = "reported" // User reported, awaiting moderation
)

// IsValid checks if the ReviewStatus is one of the defined constants.
func (s ReviewStatus) IsValid() bool {
	switch s {
	case ReviewStatusPending, ReviewStatusApproved, ReviewStatusRejected, ReviewStatusHidden, ReviewStatusReported:
		return true
	}
	return false
}

// --- Review Entity ---

// Review represents a user's review for a product or seller.
// Note: All `bson` tags have been removed from this domain entity.
// The mapping to database structures is handled by the repository implementation.
type Review struct {
	ID                primitive.ObjectID // Unique identifier for the review
	UserID            string             // ID of the user who wrote the review
	ProductID         string             // ID of the product being reviewed (e.g., ListingID)
	SellerID          string             // (Optional) ID of the seller being reviewed
	Rating            int32              // Rating given, e.g., 1-5 stars
	Comment           string             // Text content of the review
	Status            ReviewStatus       // Moderation status (e.g., pending, approved, rejected)
	ModerationComment string             // (Optional) Comment from moderator
	CreatedAt         time.Time          // Timestamp of when the review was created
	UpdatedAt         time.Time          // Timestamp of the last update
	Version           int64              // For optimistic locking
}

// NewReview creates a new review instance.
// ProductID or SellerID must be provided.
func NewReview(userID, productID, sellerID, comment string, rating int32) (*Review, error) {
	if userID == "" {
		return nil, errors.New("userID cannot be empty")
	}
	if productID == "" && sellerID == "" {
		return nil, errors.New("either productID or sellerID must be provided")
	}
	if rating < 1 || rating > 5 { // Assuming a 1-5 rating scale
		return nil, errors.New("rating must be between 1 and 5")
	}
	// Comment can be empty if allowed by business rules.

	now := time.Now().UTC()
	return &Review{
		ID:        primitive.NewObjectID(), // Generate new ID for the domain entity
		UserID:    userID,
		ProductID: productID,
		SellerID:  sellerID,
		Rating:    rating,
		Comment:   comment,
		Status:    ReviewStatusPending, // Default status
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}, nil
}

// --- ReviewFilter for Querying ---

// ReviewFilter holds parameters for querying reviews.
type ReviewFilter struct {
	Page      int32         // For pagination
	Limit     int32         // For pagination
	Status    *ReviewStatus // Optional filter by status
	MinRating *int32        // Optional filter by minimum rating
	MaxRating *int32        // Optional filter by maximum rating
	SortBy    string        // e.g., "created_at", "rating"
	SortOrder string        // "asc" or "desc"
}
