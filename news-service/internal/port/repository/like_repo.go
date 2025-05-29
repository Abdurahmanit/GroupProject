package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

type LikeRepository interface {
	AddLike(ctx context.Context, contentType string, contentID string, userID string) error
	RemoveLike(ctx context.Context, contentType string, contentID string, userID string) error
	GetLikesCount(ctx context.Context, contentType string, contentID string) (int64, error)
	HasLiked(ctx context.Context, contentType string, contentID string, userID string) (bool, error)
	DeleteByContentID(ctx context.Context, contentType string, contentID string, sessionContext mongo.SessionContext) (int64, error)
}
