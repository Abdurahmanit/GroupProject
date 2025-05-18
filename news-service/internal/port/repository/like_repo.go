package repository

import (
	"context"
)

type LikeRepository interface {
	AddLike(ctx context.Context, contentType string, contentID string, userID string) error
	RemoveLike(ctx context.Context, contentType string, contentID string, userID string) error
	GetLikesCount(ctx context.Context, contentType string, contentID string) (int64, error)
	HasLiked(ctx context.Context, contentType string, contentID string, userID string) (bool, error)
}
