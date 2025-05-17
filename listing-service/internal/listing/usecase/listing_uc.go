package usecase

import (
	"context"
	"time"

	"github.com/Abdurahmanit/GroupProject/listing-service/internal/listing/domain"
)

type ListingUsecase struct {
	repo domain.ListingRepository
}

func NewListingUsecase(repo domain.ListingRepository) *ListingUsecase {
	return &ListingUsecase{repo: repo}
}

func (uc *ListingUsecase) CreateListing(ctx context.Context, title, description string, price float64) (*domain.Listing, error) {
	listing := &domain.Listing{
		Title:       title,
		Description: description,
		Price:       price,
		Status:      domain.StatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := uc.repo.Create(ctx, listing)
	return listing, err
}

func (uc *ListingUsecase) UpdateListing(ctx context.Context, id, title, description string, price float64, status domain.ListingStatus) (*domain.Listing, error) {
	listing, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	listing.Title = title
	listing.Description = description
	listing.Price = price
	listing.Status = status
	listing.UpdatedAt = time.Now()
	err = uc.repo.Update(ctx, listing)
	return listing, err
}

func (uc *ListingUsecase) DeleteListing(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *ListingUsecase) GetListingByID(ctx context.Context, id string) (*domain.Listing, error) {
	return uc.repo.FindByID(ctx, id)
}

func (uc *ListingUsecase) SearchListings(ctx context.Context, filter domain.Filter) ([]*domain.Listing, error) {
	return uc.repo.FindByFilter(ctx, filter)
}