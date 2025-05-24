package usecase

import (
	"context"
	"time"

	"github.com/Abdurahmanit/GroupProject/listing-service/internal/listing/domain"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/platform/logger" // <--- ДОБАВИТЬ ИМПОРТ ЛОГГЕРА
)

type FavoriteUsecase struct {
	repo   domain.FavoriteRepository
	logger *logger.Logger // <--- ДОБАВЛЕНО
}

func NewFavoriteUsecase(repo domain.FavoriteRepository, log *logger.Logger) *FavoriteUsecase { // <--- ДОБАВЛЕН ЛОГГЕР
	return &FavoriteUsecase{
		repo:   repo,
		logger: log, // <--- СОХРАНЕН
	}
}

func (uc *FavoriteUsecase) AddFavorite(ctx context.Context, userID, listingID string) error {
	uc.logger.Info("FavoriteUsecase.AddFavorite: adding favorite", "user_id", userID, "listing_id", listingID)
	favorite := &domain.Favorite{
		UserID:    userID,
		ListingID: listingID,
		CreatedAt: time.Now(),
	}
	err := uc.repo.Add(ctx, favorite)
	if err != nil {
		uc.logger.Error("FavoriteUsecase.AddFavorite: failed to add favorite", "user_id", userID, "listing_id", listingID, "error", err.Error())
	}
	return err
}

func (uc *FavoriteUsecase) RemoveFavorite(ctx context.Context, userID, listingID string) error {
	uc.logger.Info("FavoriteUsecase.RemoveFavorite: removing favorite", "user_id", userID, "listing_id", listingID)
	err := uc.repo.Remove(ctx, userID, listingID)
	if err != nil {
		uc.logger.Error("FavoriteUsecase.RemoveFavorite: failed to remove favorite", "user_id", userID, "listing_id", listingID, "error", err.Error())
	}
	return err
}

func (uc *FavoriteUsecase) GetFavorites(ctx context.Context, userID string) ([]*domain.Favorite, error) {
	uc.logger.Info("FavoriteUsecase.GetFavorites: fetching favorites", "user_id", userID)
	favorites, err := uc.repo.FindByUserID(ctx, userID)
	if err != nil {
		uc.logger.Error("FavoriteUsecase.GetFavorites: failed to fetch favorites", "user_id", userID, "error", err.Error())
	}
	return favorites, err
}