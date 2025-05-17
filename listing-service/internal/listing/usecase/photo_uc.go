package usecase

import (
	"context"
	"time"

	"github.com/your-org/bike-store/listing-service/internal/listing/domain"
)

type PhotoUsecase struct {
	storage Storage
	repo    domain.ListingRepository
}

type Storage interface {
	Upload(ctx context.Context, fileName string, data []byte) (string, error)
}

func NewPhotoUsecase(storage Storage, repo domain.ListingRepository) *PhotoUsecase {
	return &PhotoUsecase{storage: storage, repo: repo}
}

func (uc *PhotoUsecase) UploadPhoto(ctx context.Context, listingID string, fileName string, data []byte) (string, error) {
	url, err := uc.storage.Upload(ctx, fileName, data)
	if err != nil {
		return "", err
	}

	listing, err := uc.repo.FindByID(ctx, listingID)
	if err != nil {
		return "", err
	}
	listing.Photos = append(listing.Photos, url)
	listing.UpdatedAt = time.Now()
	err = uc.repo.Update(ctx, listing)
	return url, err
}