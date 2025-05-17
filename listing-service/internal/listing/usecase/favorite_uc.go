package usecase

import (
	"context"
	"time"

	"github.com/your-org/bike-store/listing-service/internal/listing/domain"
)

type FavoriteUsecase struct {
	repo domain.FavoriteRepository
}

func NewFavoriteUsecase(repo domain.FavoriteRepository) *FavoriteUsecase {
	return &FavoriteUsecase{repo: repo}
}

func (uc *FavoriteUsecase) AddFavorite(ctx context.Context, userID, listingID string) error {
	favorite := &domain.Favorite{
		UserID:    userID,
		ListingID: listingID,
		CreatedAt: time.Now(),
	}
	return uc.repo.Add(ctx, favorite)
}

func (uc *FavoriteUsecase) RemoveFavorite(ctx context.Context, userID, listingID string) error {
	return uc.repo.Remove(ctx, userID, listingID)
}

func (uc *FavoriteUsecase) GetFavorites(ctx context.Context, userID string) ([]*domain.Favorite, error) {
	return uc.repo.FindByUserID(ctx, userID)
}