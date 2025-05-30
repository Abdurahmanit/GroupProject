package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Abdurahmanit/GroupProject/review-service/internal/adapter/messaging/nats" // For NATS publisher
	"github.com/Abdurahmanit/GroupProject/review-service/internal/domain"
	"github.com/Abdurahmanit/GroupProject/review-service/internal/platform/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// ReviewUsecase implements the business logic for reviews.
type ReviewUsecase struct {
	repo    domain.ReviewRepository
	natsPub *nats.Publisher // NATS publisher for events
	logger  *logger.Logger
	// adminRole string // Could be configured, e.g., "admin"
}

// NewReviewUsecase creates a new ReviewUsecase.
func NewReviewUsecase(repo domain.ReviewRepository, natsPub *nats.Publisher, log *logger.Logger) *ReviewUsecase {
	return &ReviewUsecase{
		repo:    repo,
		natsPub: natsPub,
		logger:  log.Named("ReviewUsecase"),
		// adminRole: "admin", // Default or from config
	}
}

// CreateReviewInput holds the input parameters for creating a review.
type CreateReviewInput struct {
	UserID    string
	ProductID string
	SellerID  string // Optional
	Comment   string
	Rating    int32
}

// CreateReview handles the creation of a new review.
func (uc *ReviewUsecase) CreateReview(ctx context.Context, userID, productID, sellerID, comment string, rating int32) (*domain.Review, error) {
	uc.logger.Info("Creating review",
		zap.String("user_id", userID),
		zap.String("product_id", productID),
		zap.String("seller_id", sellerID),
		zap.Int32("rating", rating))

	// Basic validation
	if userID == "" {
		return nil, fmt.Errorf("%w: userID cannot be empty", domain.ErrInvalidInput)
	}
	if productID == "" && sellerID == "" {
		return nil, fmt.Errorf("%w: productID or sellerID must be provided", domain.ErrInvalidInput)
	}
	if rating < 1 || rating > 5 {
		return nil, fmt.Errorf("%w: rating must be between 1 and 5", domain.ErrInvalidInput)
	}
	// Comment length validation could be added here or in domain.NewReview

	// Check for existing review by this user for this product/seller (if business rule applies)
	// This requires a repository method like FindByUserIDAndTargetID
	// For now, assuming the unique index in MongoDB handles this and returns domain.ErrReviewAlreadyExists

	review, err := domain.NewReview(userID, productID, sellerID, comment, rating)
	if err != nil {
		uc.logger.Error("Failed to create new domain review instance", zap.Error(err))
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidInput, err)
	}
	// Default status is set by domain.NewReview (e.g., Pending)

	err = uc.repo.Create(ctx, review)
	if err != nil {
		uc.logger.Error("Failed to save review to repository", zap.Error(err))
		if errors.Is(err, domain.ErrReviewAlreadyExists) {
			return nil, err
		}
		return nil, fmt.Errorf("%w: failed to create review: %v", domain.ErrRepository, err)
	}

	// Publish event to NATS
	eventData := map[string]interface{}{
		"review_id":  review.ID.Hex(),
		"user_id":    review.UserID,
		"product_id": review.ProductID,
		"seller_id":  review.SellerID,
		"rating":     review.Rating,
		"status":     review.Status,
		"created_at": review.CreatedAt.Format(time.RFC3339Nano),
	}
	if err := uc.natsPub.Publish(ctx, "review.created", eventData); err != nil {
		uc.logger.Warn("Failed to publish review.created event to NATS", zap.Error(err), zap.String("review_id", review.ID.Hex()))
		// Non-critical error, review is created, but event not published. Log and continue.
	}

	uc.logger.Info("Review created successfully", zap.String("review_id", review.ID.Hex()))
	return review, nil
}

// GetReview retrieves a review by its ID.
func (uc *ReviewUsecase) GetReview(ctx context.Context, reviewID primitive.ObjectID) (*domain.Review, error) {
	uc.logger.Info("Getting review by ID", zap.String("review_id", reviewID.Hex()))
	review, err := uc.repo.GetByID(ctx, reviewID)
	if err != nil {
		uc.logger.Error("Failed to get review from repository", zap.Error(err), zap.String("review_id", reviewID.Hex()))
		return nil, err // repo.GetByID should return domain.ErrNotFound
	}
	return review, nil
}

// UpdateReview allows a user to update their own review (comment/rating).
// Only the author of the review can update it, and only if it's not heavily moderated.
func (uc *ReviewUsecase) UpdateReview(ctx context.Context, reviewID primitive.ObjectID, userID string, rating *int32, comment *string) (*domain.Review, error) {
	uc.logger.Info("Updating review",
		zap.String("review_id", reviewID.Hex()),
		zap.String("user_id", userID))

	review, err := uc.repo.GetByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}

	if review.UserID != userID {
		uc.logger.Warn("User forbidden to update review", zap.String("review_id", reviewID.Hex()), zap.String("review_author", review.UserID), zap.String("requesting_user", userID))
		return nil, domain.ErrForbidden
	}

	// Business rule: Maybe only 'approved' or 'pending' reviews can be updated by the user.
	// if review.Status == domain.ReviewStatusRejected || review.Status == domain.ReviewStatusHidden {
	// 	return nil, fmt.Errorf("%w: cannot update a review that is '%s'", domain.ErrForbidden, review.Status)
	// }

	updated := false
	if rating != nil {
		if *rating < 1 || *rating > 5 {
			return nil, fmt.Errorf("%w: rating must be between 1 and 5", domain.ErrInvalidInput)
		}
		if review.Rating != *rating {
			review.Rating = *rating
			updated = true
		}
	}
	if comment != nil {
		// Add comment length validation if needed
		if review.Comment != *comment {
			review.Comment = *comment
			updated = true
		}
	}

	if !updated {
		uc.logger.Info("No changes detected for review update", zap.String("review_id", reviewID.Hex()))
		return review, nil // Return existing review if no changes
	}

	review.UpdatedAt = time.Now().UTC()
	review.Version++ // Increment version for optimistic locking

	// If a user updates their review, it might need to go back to pending status for re-moderation.
	// review.Status = domain.ReviewStatusPending

	err = uc.repo.Update(ctx, review)
	if err != nil {
		return nil, err
	}

	// Publish event
	eventData := map[string]interface{}{
		"review_id":  review.ID.Hex(),
		"user_id":    review.UserID,
		"product_id": review.ProductID,
		"updated_at": review.UpdatedAt.Format(time.RFC3339Nano),
	}
	uc.natsPub.Publish(ctx, "review.updated", eventData) // Error handling for NATS as in CreateReview

	uc.logger.Info("Review updated successfully", zap.String("review_id", review.ID.Hex()))
	return review, nil
}

// DeleteReview allows a user to delete their own review.
func (uc *ReviewUsecase) DeleteReview(ctx context.Context, reviewID primitive.ObjectID, userID string) error {
	uc.logger.Info("Deleting review", zap.String("review_id", reviewID.Hex()), zap.String("user_id", userID))

	review, err := uc.repo.GetByID(ctx, reviewID)
	if err != nil {
		return err
	}

	if review.UserID != userID { // Add admin role check here if admins can delete any review
		uc.logger.Warn("User forbidden to delete review", zap.String("review_id", reviewID.Hex()), zap.String("review_author", review.UserID), zap.String("requesting_user", userID))
		return domain.ErrForbidden
	}

	err = uc.repo.Delete(ctx, reviewID)
	if err != nil {
		return err
	}

	// Publish event
	eventData := map[string]interface{}{
		"review_id":  reviewID.Hex(),
		"user_id":    userID,           // User who performed the delete
		"product_id": review.ProductID, // Include product ID for context
		"deleted_at": time.Now().UTC().Format(time.RFC3339Nano),
	}
	uc.natsPub.Publish(ctx, "review.deleted", eventData)

	uc.logger.Info("Review deleted successfully", zap.String("review_id", reviewID.Hex()))
	return nil
}

// ListReviewsByProduct retrieves reviews for a product with pagination and status filter.
func (uc *ReviewUsecase) ListReviewsByProduct(ctx context.Context, productID string, page, limit int32, statusFilter *string) ([]*domain.Review, int64, error) {
	uc.logger.Info("Listing reviews by product", zap.String("product_id", productID), zap.Int32("page", page), zap.Int32("limit", limit), zap.Any("status_filter", statusFilter))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10 // Default limit
	} else if limit > 100 {
		limit = 100 // Max limit
	}

	filter := domain.ReviewFilter{
		Page:  page,
		Limit: limit,
	}
	if statusFilter != nil {
		s := domain.ReviewStatus(*statusFilter)
		if !s.IsValid() {
			return nil, 0, fmt.Errorf("%w: invalid status filter value '%s'", domain.ErrInvalidInput, *statusFilter)
		}
		filter.Status = &s
	} else {
		// Default to only showing approved reviews for public listings
		approvedStatus := domain.ReviewStatusApproved
		filter.Status = &approvedStatus
	}

	return uc.repo.FindByProductID(ctx, productID, filter)
}

// ListReviewsByUser retrieves reviews by a user with pagination.
func (uc *ReviewUsecase) ListReviewsByUser(ctx context.Context, userID string, page, limit int32) ([]*domain.Review, int64, error) {
	uc.logger.Info("Listing reviews by user", zap.String("user_id", userID), zap.Int32("page", page), zap.Int32("limit", limit))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}
	filter := domain.ReviewFilter{Page: page, Limit: limit} // No status filter by default for user's own reviews
	return uc.repo.FindByUserID(ctx, userID, filter)
}

// ModerateReview allows an admin to change the status of a review.
// Assumes admin authorization (role check) is handled by the gRPC auth interceptor or here.
func (uc *ReviewUsecase) ModerateReview(ctx context.Context, reviewID primitive.ObjectID, adminUserID string, newStatus domain.ReviewStatus, moderationComment string) (*domain.Review, error) {
	uc.logger.Info("Moderating review",
		zap.String("review_id", reviewID.Hex()),
		zap.String("admin_user_id", adminUserID),
		zap.String("new_status", string(newStatus)))

	// Here, you might fetch the adminUser to verify their role if not done by interceptor.
	// For simplicity, assuming adminUserID is validated as an admin.

	if !newStatus.IsValid() {
		return nil, fmt.Errorf("%w: invalid new status '%s'", domain.ErrInvalidInput, newStatus)
	}

	review, err := uc.repo.GetByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}

	if review.Status == newStatus && review.ModerationComment == moderationComment {
		uc.logger.Info("No change in status or moderation comment for review", zap.String("review_id", reviewID.Hex()))
		return review, nil // No actual change
	}

	oldStatus := review.Status
	review.Status = newStatus
	review.ModerationComment = moderationComment
	review.UpdatedAt = time.Now().UTC()
	review.Version++

	err = uc.repo.Update(ctx, review)
	if err != nil {
		return nil, err
	}

	// Publish event
	eventData := map[string]interface{}{
		"review_id":          review.ID.Hex(),
		"moderator_id":       adminUserID,
		"product_id":         review.ProductID,
		"old_status":         oldStatus,
		"new_status":         newStatus,
		"moderation_comment": moderationComment,
		"moderated_at":       review.UpdatedAt.Format(time.RFC3339Nano),
	}
	uc.natsPub.Publish(ctx, "review.moderated", eventData)

	uc.logger.Info("Review moderated successfully", zap.String("review_id", review.ID.Hex()), zap.String("new_status", string(newStatus)))
	return review, nil
}

// GetProductAverageRating calculates and returns the average rating for a product.
func (uc *ReviewUsecase) GetProductAverageRating(ctx context.Context, productID string) (float64, int32, error) {
	uc.logger.Info("Getting average rating for product", zap.String("product_id", productID))
	if productID == "" {
		return 0, 0, fmt.Errorf("%w: productID cannot be empty", domain.ErrInvalidInput)
	}
	return uc.repo.GetAverageRating(ctx, productID)
}
