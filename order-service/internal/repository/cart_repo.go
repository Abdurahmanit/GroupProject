package repository

import (
	"context"
	"time"

	"github.com/Abdurahmanit/GroupProject/order-service/internal/domain/entity"
)

type CartRepository interface {
	GetByUserID(ctx context.Context, userID string) (*entity.Cart, error)
	Save(ctx context.Context, cart *entity.Cart, ttl time.Duration) error
	DeleteByUserID(ctx context.Context, userID string) error
}
