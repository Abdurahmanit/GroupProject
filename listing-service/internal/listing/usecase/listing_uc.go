package usecase

import (
	"context"
	"errors" // Для кастомных ошибок
	"time"
	"fmt"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/listing/domain"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/platform/logger" // <--- ДОБАВИТЬ ИМПОРТ ЛОГГЕРА
)

// Определим ошибки для usecase слоя
var (
	ErrListingNotFound = errors.New("listing not found")
	ErrForbidden       = errors.New("user not authorized to perform this action")
)

type ListingUsecase struct {
	repo   domain.ListingRepository
	logger *logger.Logger // <--- ДОБАВЛЕНО
}

func NewListingUsecase(repo domain.ListingRepository, log *logger.Logger) *ListingUsecase { // <--- ДОБАВЛЕН ЛОГГЕР
	return &ListingUsecase{
		repo:   repo,
		logger: log, // <--- СОХРАНЕН
	}
}

// CreateListing теперь принимает userID и categoryID
func (uc *ListingUsecase) CreateListing(ctx context.Context, userID, categoryID, title, description string, price float64) (*domain.Listing, error) {
	uc.logger.Info("ListingUsecase.CreateListing: creating new listing",
		"user_id", userID, "category_id", categoryID, "title", title)

	listing := &domain.Listing{
		UserID:      userID, // <--- СОХРАНЯЕМ
		CategoryID:  categoryID, // <--- СОХРАНЯЕМ
		Title:       title,
		Description: description,
		Price:       price,
		Status:      domain.StatusActive, // Убедись, что StatusActive определен в domain
		Photos:      []string{},          // Инициализируем пустым слайсом
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := uc.repo.Create(ctx, listing)
	if err != nil {
		uc.logger.Error("ListingUsecase.CreateListing: failed to create listing", "error", err.Error(), "user_id", userID)
		return nil, err
	}
	return listing, nil
}

// UpdateListing теперь принимает userID для авторизации и categoryID
func (uc *ListingUsecase) UpdateListing(ctx context.Context, id, userID, categoryID, title, description string, price float64, status domain.ListingStatus) (*domain.Listing, error) {
	uc.logger.Info("ListingUsecase.UpdateListing: updating listing",
		"listing_id", id, "user_id_performing_action", userID)

	listing, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		uc.logger.Error("ListingUsecase.UpdateListing: failed to find listing", "listing_id", id, "error", err.Error())
		if errors.Is(err, domain.ErrListingNotFound) { // Предполагаем, что репозиторий возвращает такую ошибку
			return nil, ErrListingNotFound
		}
		return nil, err
	}
	if listing == nil { // Дополнительная проверка
		uc.logger.Warn("ListingUsecase.UpdateListing: listing not found by ID", "listing_id", id)
		return nil, ErrListingNotFound
	}

	// Авторизация: только владелец может обновлять
	if listing.UserID != userID {
		uc.logger.Warn("ListingUsecase.UpdateListing: forbidden to update listing",
			"listing_id", id, "listing_owner_id", listing.UserID, "user_id_performing_action", userID)
		return nil, ErrForbidden
	}

	// Обновляем поля, если они переданы (проверка на пустые строки/значения по умолчанию может быть добавлена)
	if title != "" {
		listing.Title = title
	}
	if description != "" {
		listing.Description = description
	}
	if price > 0 { // Пример: цена должна быть больше 0 для обновления
		listing.Price = price
	}
	if categoryID != "" {
		listing.CategoryID = categoryID
	}
	if status != "" && status != listing.Status { // Обновляем статус, если он передан и отличается
		listing.Status = status
	}
	listing.UpdatedAt = time.Now()

	err = uc.repo.Update(ctx, listing)
	if err != nil {
		uc.logger.Error("ListingUsecase.UpdateListing: failed to update listing in repo", "listing_id", id, "error", err.Error())
		return nil, err
	}
	return listing, nil
}

// DeleteListing теперь принимает userID для авторизации
func (uc *ListingUsecase) DeleteListing(ctx context.Context, id, userID string) error {
	uc.logger.Info("ListingUsecase.DeleteListing: deleting listing",
		"listing_id", id, "user_id_performing_action", userID)

	listing, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		uc.logger.Error("ListingUsecase.DeleteListing: failed to find listing", "listing_id", id, "error", err.Error())
		if errors.Is(err, domain.ErrListingNotFound) {
			return ErrListingNotFound
		}
		return err
	}
    if listing == nil {
		uc.logger.Warn("ListingUsecase.DeleteListing: listing not found by ID", "listing_id", id)
		return ErrListingNotFound
	}

	// Авторизация: только владелец может удалять
	if listing.UserID != userID {
		uc.logger.Warn("ListingUsecase.DeleteListing: forbidden to delete listing",
			"listing_id", id, "listing_owner_id", listing.UserID, "user_id_performing_action", userID)
		return ErrForbidden
	}

	err = uc.repo.Delete(ctx, id)
	if err != nil {
		uc.logger.Error("ListingUsecase.DeleteListing: failed to delete listing in repo", "listing_id", id, "error", err.Error())
	}
	return err
}

func (uc *ListingUsecase) GetListingByID(ctx context.Context, id string) (*domain.Listing, error) {
	uc.logger.Info("ListingUsecase.GetListingByID: fetching listing", "listing_id", id)
	listing, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		uc.logger.Warn("ListingUsecase.GetListingByID: failed to find listing", "listing_id", id, "error", err.Error())
		if errors.Is(err, domain.ErrListingNotFound) {
			return nil, ErrListingNotFound
		}
		return nil, err
	}
    if listing == nil {
		uc.logger.Warn("ListingUsecase.GetListingByID: listing not found by ID", "listing_id", id)
		return nil, ErrListingNotFound
	}
	return listing, nil
}

// SearchListings теперь возвращает (listings, total, error)
func (uc *ListingUsecase) SearchListings(ctx context.Context, filter domain.Filter) ([]*domain.Listing, int64, error) {
	uc.logger.Info("ListingUsecase.SearchListings: searching listings", "filter", fmt.Sprintf("%+v", filter))
	// Предполагаем, что FindByFilter в репозитории теперь возвращает (listings, total, error)
	// Если нет, тебе нужно будет либо изменить репозиторий, либо сделать два запроса: один для данных, другой для count(*).
	listings, total, err := uc.repo.FindByFilter(ctx, filter)
	if err != nil {
		uc.logger.Error("ListingUsecase.SearchListings: failed to search listings", "filter", fmt.Sprintf("%+v", filter), "error", err.Error())
		return nil, 0, err
	}
	return listings, total, nil
}

// UpdateListingStatus - новый метод
func (uc *ListingUsecase) UpdateListingStatus(ctx context.Context, id, userID string, status domain.ListingStatus) (*domain.Listing, error) {
	uc.logger.Info("ListingUsecase.UpdateListingStatus: updating listing status",
		"listing_id", id, "user_id_performing_action", userID, "new_status", string(status))

	listing, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		uc.logger.Error("ListingUsecase.UpdateListingStatus: failed to find listing", "listing_id", id, "error", err.Error())
		if errors.Is(err, domain.ErrListingNotFound) {
			return nil, ErrListingNotFound
		}
		return nil, err
	}
    if listing == nil {
		uc.logger.Warn("ListingUsecase.UpdateListingStatus: listing not found by ID", "listing_id", id)
		return nil, ErrListingNotFound
	}

	// Авторизация: только владелец может обновлять статус
	if listing.UserID != userID {
		uc.logger.Warn("ListingUsecase.UpdateListingStatus: forbidden to update listing status",
			"listing_id", id, "listing_owner_id", listing.UserID, "user_id_performing_action", userID)
		return nil, ErrForbidden
	}

	if status == "" { // Нельзя установить пустой статус
		uc.logger.Warn("ListingUsecase.UpdateListingStatus: attempt to set empty status", "listing_id", id)
		return nil, errors.New("status cannot be empty") // Или более специфичная ошибка
	}

	listing.Status = status
	listing.UpdatedAt = time.Now()

	err = uc.repo.Update(ctx, listing) // Используем тот же Update, что и для полного обновления
	if err != nil {
		uc.logger.Error("ListingUsecase.UpdateListingStatus: failed to update listing status in repo", "listing_id", id, "error", err.Error())
		return nil, err
	}
	return listing, nil
}