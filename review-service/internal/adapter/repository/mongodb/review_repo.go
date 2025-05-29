package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Abdurahmanit/GroupProject/review-service/internal/domain"
	"github.com/Abdurahmanit/GroupProject/review-service/internal/platform/logger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	zap "go.uber.org/zap"
)

const reviewCollectionName = "reviews"

// ReviewRepository implements the domain.ReviewRepository interface using MongoDB.
type ReviewRepository struct {
	collection *mongo.Collection
	logger     *logger.Logger
}

// NewReviewRepository creates a new MongoDB review repository.
func NewReviewRepository(db *mongo.Database, log *logger.Logger) (*ReviewRepository, error) {
	collection := db.Collection(reviewCollectionName)

	// Define indexes
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "product_id", Value: 1}, {Key: "status", Value: 1}}}, // For querying reviews by product and status
		{Keys: bson.D{{Key: "user_id", Value: 1}}},                               // For querying reviews by user
		{Keys: bson.D{{Key: "product_id", Value: 1}, {Key: "user_id", Value: 1}}, Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{"seller_id": bson.M{"$exists": false}})}, // Unique review per user per product
		{Keys: bson.D{{Key: "seller_id", Value: 1}, {Key: "user_id", Value: 1}}, Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{"product_id": bson.M{"$exists": false}})}, // Unique review per user per seller (if applicable)
		{Keys: bson.D{{Key: "status", Value: 1}}}, // For querying by status (e.g., pending moderation)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Error("Failed to create indexes for reviews collection", zap.Error(err))
		// Don't necessarily fail startup, as indexes might already exist or be created manually.
		// return nil, fmt.Errorf("failed to create indexes for %s: %w", reviewCollectionName, err)
	} else {
		log.Info("Successfully ensured indexes for reviews collection")
	}

	return &ReviewRepository{
		collection: collection,
		logger:     log.Named("ReviewRepository"),
	}, nil
}

// Create inserts a new review into the database.
func (r *ReviewRepository) Create(ctx context.Context, review *domain.Review) error {
	r.logger.Info("Creating review in DB", zap.String("product_id", review.ProductID), zap.String("user_id", review.UserID))

	doc, err := fromDomainReview(review)
	if err != nil {
		r.logger.Error("Failed to convert domain.Review to document for Create", zap.Error(err))
		return err
	}
	if doc.ID.IsZero() { // Ensure ID is set if not already
		doc.ID = primitive.NewObjectID()
	}
	review.ID = doc.ID // Update domain entity with generated/confirmed ID

	now := time.Now().UTC()
	doc.CreatedAt = now
	doc.UpdatedAt = now
	review.CreatedAt = now // Ensure domain model has timestamps
	review.UpdatedAt = now

	_, err = r.collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			r.logger.Warn("Duplicate key error on review creation", zap.Error(err))
			return domain.ErrReviewAlreadyExists // Use domain-specific error
		}
		r.logger.Error("Failed to insert review into DB", zap.Error(err))
		return fmt.Errorf("db insert failed: %w", err)
	}
	r.logger.Info("Review created successfully in DB", zap.String("review_id", doc.ID.Hex()))
	return nil
}

// GetByID retrieves a review by its ID.
func (r *ReviewRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Review, error) {
	r.logger.Debug("Getting review by ID from DB", zap.String("review_id", id.Hex()))
	var doc reviewDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			r.logger.Warn("Review not found in DB", zap.String("review_id", id.Hex()))
			return nil, domain.ErrNotFound
		}
		r.logger.Error("Failed to get review by ID from DB", zap.Error(err), zap.String("review_id", id.Hex()))
		return nil, fmt.Errorf("db findone failed: %w", err)
	}
	return doc.toDomainReview(), nil
}

// Update modifies an existing review in the database.
func (r *ReviewRepository) Update(ctx context.Context, review *domain.Review) error {
	r.logger.Info("Updating review in DB", zap.String("review_id", review.ID.Hex()))
	if review.ID.IsZero() {
		return errors.New("cannot update review without ID")
	}

	doc, err := fromDomainReview(review)
	if err != nil {
		r.logger.Error("Failed to convert domain.Review to document for Update", zap.Error(err))
		return err
	}
	doc.UpdatedAt = time.Now().UTC()
	review.UpdatedAt = doc.UpdatedAt // Sync domain model

	// Construct update document to only set fields that are typically updatable
	updatePayload := bson.M{
		"$set": bson.M{
			"rating":             doc.Rating,
			"comment":            doc.Comment,
			"status":             doc.Status,
			"moderation_comment": doc.ModerationComment,
			"updated_at":         doc.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": doc.ID}, updatePayload)
	if err != nil {
		r.logger.Error("Failed to update review in DB", zap.Error(err), zap.String("review_id", doc.ID.Hex()))
		return fmt.Errorf("db update failed: %w", err)
	}
	if result.MatchedCount == 0 {
		r.logger.Warn("Review not found for update in DB", zap.String("review_id", doc.ID.Hex()))
		return domain.ErrNotFound
	}
	r.logger.Info("Review updated successfully in DB", zap.String("review_id", doc.ID.Hex()))
	return nil
}

// Delete removes a review from the database.
func (r *ReviewRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	r.logger.Info("Deleting review from DB", zap.String("review_id", id.Hex()))
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		r.logger.Error("Failed to delete review from DB", zap.Error(err), zap.String("review_id", id.Hex()))
		return fmt.Errorf("db delete failed: %w", err)
	}
	if result.DeletedCount == 0 {
		r.logger.Warn("Review not found for deletion in DB", zap.String("review_id", id.Hex()))
		return domain.ErrNotFound
	}
	r.logger.Info("Review deleted successfully from DB", zap.String("review_id", id.Hex()))
	return nil
}

// FindByProductID retrieves reviews for a specific product, with pagination and optional status filter.
func (r *ReviewRepository) FindByProductID(ctx context.Context, productID string, filter domain.ReviewFilter) ([]*domain.Review, int64, error) {
	r.logger.Debug("Finding reviews by product_id from DB", zap.String("product_id", productID), zap.Any("filter", filter))

	mongoQuery := bson.M{"product_id": productID}
	if filter.Status != nil {
		mongoQuery["status"] = *filter.Status
	}

	findOptions := options.Find()
	if filter.Limit > 0 {
		findOptions.SetLimit(int64(filter.Limit))
		if filter.Page > 0 {
			findOptions.SetSkip(int64(filter.Page-1) * int64(filter.Limit))
		}
	}
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}}) // Newest first

	cursor, err := r.collection.Find(ctx, mongoQuery, findOptions)
	if err != nil {
		r.logger.Error("Failed to find reviews by product_id from DB", zap.Error(err), zap.String("product_id", productID))
		return nil, 0, fmt.Errorf("db find failed: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []*reviewDocument
	if err = cursor.All(ctx, &docs); err != nil {
		r.logger.Error("Failed to decode reviews by product_id from DB", zap.Error(err))
		return nil, 0, fmt.Errorf("db cursor all failed: %w", err)
	}

	domainReviews := make([]*domain.Review, len(docs))
	for i, doc := range docs {
		domainReviews[i] = doc.toDomainReview()
	}

	total, err := r.collection.CountDocuments(ctx, mongoQuery)
	if err != nil {
		r.logger.Error("Failed to count reviews by product_id from DB", zap.Error(err))
		return nil, 0, fmt.Errorf("db count failed: %w", err)
	}

	return domainReviews, total, nil
}

// FindByUserID retrieves reviews written by a specific user, with pagination.
func (r *ReviewRepository) FindByUserID(ctx context.Context, userID string, filter domain.ReviewFilter) ([]*domain.Review, int64, error) {
	r.logger.Debug("Finding reviews by user_id from DB", zap.String("user_id", userID), zap.Any("filter", filter))

	mongoQuery := bson.M{"user_id": userID}
	// Could add status filter here too if needed for "my reviews" page

	findOptions := options.Find()
	if filter.Limit > 0 {
		findOptions.SetLimit(int64(filter.Limit))
		if filter.Page > 0 {
			findOptions.SetSkip(int64(filter.Page-1) * int64(filter.Limit))
		}
	}
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, mongoQuery, findOptions)
	if err != nil {
		r.logger.Error("Failed to find reviews by user_id from DB", zap.Error(err), zap.String("user_id", userID))
		return nil, 0, fmt.Errorf("db find failed: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []*reviewDocument
	if err = cursor.All(ctx, &docs); err != nil {
		r.logger.Error("Failed to decode reviews by user_id from DB", zap.Error(err))
		return nil, 0, fmt.Errorf("db cursor all failed: %w", err)
	}

	domainReviews := make([]*domain.Review, len(docs))
	for i, doc := range docs {
		domainReviews[i] = doc.toDomainReview()
	}

	total, err := r.collection.CountDocuments(ctx, mongoQuery)
	if err != nil {
		r.logger.Error("Failed to count reviews by user_id from DB", zap.Error(err))
		return nil, 0, fmt.Errorf("db count failed: %w", err)
	}

	return domainReviews, total, nil
}

// GetAverageRating calculates the average rating for a product.
// This might be better handled by an aggregation pipeline or a separate denormalized field.
func (r *ReviewRepository) GetAverageRating(ctx context.Context, productID string) (float64, int32, error) {
	r.logger.Debug("Calculating average rating for product_id", zap.String("product_id", productID))

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "product_id", Value: productID},
			{Key: "status", Value: domain.ReviewStatusApproved}, // Only consider approved reviews for average
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$product_id"},
			{Key: "average_rating", Value: bson.D{{Key: "$avg", Value: "$rating"}}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		r.logger.Error("Failed to aggregate average rating", zap.Error(err), zap.String("product_id", productID))
		return 0, 0, fmt.Errorf("db aggregate failed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		AverageRating float64 `bson:"average_rating"`
		Count         int32   `bson:"count"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		r.logger.Error("Failed to decode average rating aggregation result", zap.Error(err))
		return 0, 0, fmt.Errorf("db cursor all for aggregate failed: %w", err)
	}

	if len(results) == 0 {
		return 0, 0, nil // No approved reviews found for this product
	}

	return results[0].AverageRating, results[0].Count, nil
}
