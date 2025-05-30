package domain

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrNotFound            = errors.New("entity not found")
	ErrForbidden           = errors.New("action forbidden")
	ErrInvalidInput        = errors.New("invalid input data")
	ErrReviewAlreadyExists = errors.New("review already exists for this user and target")
	ErrOptimisticLock      = errors.New("optimistic lock conflict: data was modified by another process")
	ErrRepository          = errors.New("repository error")
)

type ReviewStatus string

const (
	ReviewStatusPending  ReviewStatus = "pending"
	ReviewStatusApproved ReviewStatus = "approved"
	ReviewStatusRejected ReviewStatus = "rejected"
	ReviewStatusHidden   ReviewStatus = "hidden"   // Hidden by admin, not deleted
	ReviewStatusReported ReviewStatus = "reported" // User reported, awaiting moderation
)

func (s ReviewStatus) IsValid() bool {
	switch s {
	case ReviewStatusPending, ReviewStatusApproved, ReviewStatusRejected, ReviewStatusHidden, ReviewStatusReported:
		return true
	}
	return false
}

type Review struct {
	ID                primitive.ObjectID
	UserID            string
	ProductID         string
	SellerID          string
	Rating            int32
	Comment           string
	Status            ReviewStatus
	ModerationComment string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Version           int64
}

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

type ReviewFilter struct {
	Page      int32
	Limit     int32
	Status    *ReviewStatus
	MinRating *int32
	MaxRating *int32
	SortBy    string
	SortOrder string
}
