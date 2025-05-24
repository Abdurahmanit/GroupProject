package usecase

import (
	"context"
	"errors" // Для кастомных ошибок
	"time"

	"github.com/Abdurahmanit/GroupProject/listing-service/internal/listing/domain"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/platform/logger" // <--- ДОБАВИТЬ ИМПОРТ ЛОГГЕРА
)

type PhotoUsecase struct {
	storage domain.Storage // Интерфейс Storage остается
	repo    domain.ListingRepository
	logger  *logger.Logger // <--- ДОБАВЛЕНО
}


func NewPhotoUsecase(storage domain.Storage, repo domain.ListingRepository, log *logger.Logger) *PhotoUsecase { // <--- ДОБАВЛЕН ЛОГГЕР
	return &PhotoUsecase{
		storage: storage,
		repo:    repo,
		logger:  log, // <--- СОХРАНЕН
	}
}

// UploadPhoto теперь принимает userID для авторизации
func (uc *PhotoUsecase) UploadPhoto(ctx context.Context, listingID, userID, fileName string, data []byte) (string, error) {
	uc.logger.Info("PhotoUsecase.UploadPhoto: uploading photo",
		"listing_id", listingID, "user_id_performing_action", userID, "filename", fileName)

	listing, err := uc.repo.FindByID(ctx, listingID)
	if err != nil {
		uc.logger.Error("PhotoUsecase.UploadPhoto: failed to find listing", "listing_id", listingID, "error", err.Error())
		if errors.Is(err, domain.ErrListingNotFound) { // Предполагаем, что репозиторий возвращает такую ошибку
			return "", ErrListingNotFound // Используем ошибку usecase-уровня
		}
		return "", err
	}
    if listing == nil {
		uc.logger.Warn("PhotoUsecase.UploadPhoto: listing not found by ID", "listing_id", listingID)
		return "", ErrListingNotFound
	}

	// Авторизация: только владелец может загружать фото к объявлению
	if listing.UserID != userID {
		uc.logger.Warn("PhotoUsecase.UploadPhoto: forbidden to upload photo",
			"listing_id", listingID, "listing_owner_id", listing.UserID, "user_id_performing_action", userID)
		return "", ErrForbidden // Используем ошибку usecase-уровня
	}

	url, err := uc.storage.Upload(ctx, fileName, data) // fileName должен быть уникальным или генерироваться хранилищем
	if err != nil {
		uc.logger.Error("PhotoUsecase.UploadPhoto: storage upload failed", "listing_id", listingID, "filename", fileName, "error", err.Error())
		return "", err
	}

	// Обновляем список фото в объявлении
	if listing.Photos == nil {
		listing.Photos = []string{}
	}
	listing.Photos = append(listing.Photos, url)
	listing.UpdatedAt = time.Now()

	err = uc.repo.Update(ctx, listing) // Обновляем объявление в репозитории
	if err != nil {
		uc.logger.Error("PhotoUsecase.UploadPhoto: failed to update listing after photo upload", "listing_id", listingID, "error", err.Error())
		// Здесь может потребоваться логика отката загруженного файла из storage, если обновление БД не удалось (сложно)
		return "", err
	}
	return url, nil
}