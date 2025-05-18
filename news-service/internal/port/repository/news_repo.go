package repository

import (
	"context"

	"github.com/Abdurahmanit/GroupProject/news-service/internal/entity"
)

type NewsRepository interface {
	Create(ctx context.Context, news *entity.News) (string, error)
	GetByID(ctx context.Context, id string) (*entity.News, error)
	Update(ctx context.Context, news *entity.News) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page, pageSize int, filter map[string]interface{}) ([]*entity.News, int, error)
}
