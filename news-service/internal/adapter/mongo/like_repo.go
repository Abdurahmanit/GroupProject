package mongo

import (
	"context"
	"fmt"

	"github.com/Abdurahmanit/GroupProject/news-service/internal/port/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const likesCollectionName = "likes"

type LikeMongoRepository struct {
	db *mongo.Database
}

func NewLikeMongoRepository(client *mongo.Client, dbName string) repository.LikeRepository {
	return &LikeMongoRepository{
		db: client.Database(dbName),
	}
}

type likeDocument struct {
	ContentType string `bson:"content_type"`
	ContentID   string `bson:"content_id"`
	UserID      string `bson:"user_id"`
}

func (r *LikeMongoRepository) AddLike(ctx context.Context, contentType string, contentID string, userID string) error {
	doc := likeDocument{
		ContentType: contentType,
		ContentID:   contentID,
		UserID:      userID,
	}

	filter := bson.M{
		"content_type": doc.ContentType,
		"content_id":   doc.ContentID,
		"user_id":      doc.UserID,
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.db.Collection(likesCollectionName).UpdateOne(ctx, filter, bson.M{"$setOnInsert": doc}, opts)
	if err != nil {
		return fmt.Errorf("failed to add like in mongo: %w", err)
	}
	return nil
}

func (r *LikeMongoRepository) RemoveLike(ctx context.Context, contentType string, contentID string, userID string) error {
	filter := bson.M{
		"content_type": contentType,
		"content_id":   contentID,
		"user_id":      userID,
	}
	res, err := r.db.Collection(likesCollectionName).DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to remove like from mongo: %w", err)
	}
	if res.DeletedCount == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *LikeMongoRepository) GetLikesCount(ctx context.Context, contentType string, contentID string) (int64, error) {
	filter := bson.M{
		"content_type": contentType,
		"content_id":   contentID,
	}
	count, err := r.db.Collection(likesCollectionName).CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to get likes count from mongo: %w", err)
	}
	return count, nil
}

func (r *LikeMongoRepository) HasLiked(ctx context.Context, contentType string, contentID string, userID string) (bool, error) {
	filter := bson.M{
		"content_type": contentType,
		"content_id":   contentID,
		"user_id":      userID,
	}
	count, err := r.db.Collection(likesCollectionName).CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check if liked from mongo: %w", err)
	}
	return count > 0, nil
}

func (r *LikeMongoRepository) DeleteByContentID(ctx context.Context, contentType string, contentID string, sessionContext mongo.SessionContext) (int64, error) {
	targetCtx := ctx
	if sessionContext != nil {
		targetCtx = sessionContext
	}
	filter := bson.M{
		"content_type": contentType,
		"content_id":   contentID,
	}
	res, err := r.db.Collection(likesCollectionName).DeleteMany(targetCtx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to delete likes by content_id from mongo: %w", err)
	}
	return res.DeletedCount, nil
}
