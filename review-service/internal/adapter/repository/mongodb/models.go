package mongodb

import (
	"errors"
	"time"

	"github.com/Abdurahmanit/GroupProject/review-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type reviewDocument struct {
	ID                primitive.ObjectID  `bson:"_id,omitempty"`
	UserID            string              `bson:"user_id"`
	ProductID         string              `bson:"product_id"`
	SellerID          string              `bson:"seller_id,omitempty"`
	Rating            int32               `bson:"rating"`
	Comment           string              `bson:"comment"`
	Status            domain.ReviewStatus `bson:"status"`
	ModerationComment string              `bson:"moderation_comment,omitempty"` // Comment from moderator
	CreatedAt         time.Time           `bson:"created_at"`
	UpdatedAt         time.Time           `bson:"updated_at"`
	Version           int64               `bson:"version"`
}

// toDomainReview converts a reviewDocument from MongoDB to a domain.Review entity.
func (doc *reviewDocument) toDomainReview() *domain.Review {
	if doc == nil {
		return nil
	}
	return &domain.Review{
		ID:                doc.ID,
		UserID:            doc.UserID,
		ProductID:         doc.ProductID,
		SellerID:          doc.SellerID,
		Rating:            doc.Rating,
		Comment:           doc.Comment,
		Status:            doc.Status,
		ModerationComment: doc.ModerationComment,
		CreatedAt:         doc.CreatedAt,
		UpdatedAt:         doc.UpdatedAt,
	}
}

func fromDomainReview(review *domain.Review) (*reviewDocument, error) {
	if review == nil {
		return nil, errors.New("cannot convert nil domain.Review to reviewDocument")
	}

	docID := review.ID
	if docID.IsZero() {
		docID = primitive.NewObjectID()
	}

	return &reviewDocument{
		ID:                docID,
		UserID:            review.UserID,
		ProductID:         review.ProductID,
		SellerID:          review.SellerID,
		Rating:            review.Rating,
		Comment:           review.Comment,
		Status:            review.Status,
		ModerationComment: review.ModerationComment,
		CreatedAt:         review.CreatedAt,
		UpdatedAt:         review.UpdatedAt,
	}, nil
}
