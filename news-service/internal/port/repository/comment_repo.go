package repository

import (
	"context"

	"github.com/Abdurahmanit/GroupProject/news-service/internal/entity"
	"go.mongodb.org/mongo-driver/mongo"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *entity.Comment) (string, error)
	GetByID(ctx context.Context, id string) (*entity.Comment, error)
	GetByNewsID(ctx context.Context, newsID string, page, pageSize int) ([]*entity.Comment, int, error)
	Update(ctx context.Context, comment *entity.Comment) error
	Delete(ctx context.Context, id string) error
	DeleteByNewsID(ctx context.Context, newsID string, sessionContext mongo.SessionContext) (int64, error)
}
