package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/listing/domain"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/platform/logger" // Предполагаем, что логгер передается
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options" // Для опций поиска
)

// Определим специфичные для репозитория ошибки (уже были в предыдущей версии)
var (
	ErrFavoriteAlreadyExistsDB = errors.New("database: favorite already exists for this user and listing")
	ErrFavoriteNotFoundDB      = errors.New("database: favorite not found")
)
// Эти ошибки уже должны быть определены в этом пакете или в общем месте для ошибок БД.

type FavoriteRepository struct {
	collection *mongo.Collection
	logger     *logger.Logger
}

// NewFavoriteRepository теперь принимает логгер
func NewFavoriteRepository(db *mongo.Database, log *logger.Logger) *FavoriteRepository {
	// Рекомендуется создать уникальный индекс в MongoDB для предотвращения дубликатов
	// db.collection("favorites").createIndex({ "user_id": 1, "listing_id": 1 }, { unique: true })
	// Эту операцию лучше выполнять один раз при инициализации приложения или через миграции.
	return &FavoriteRepository{
		collection: db.Collection("favorites"),
		logger:     log,
	}
}

func (r *FavoriteRepository) Add(ctx context.Context, favorite *domain.Favorite) error {
	r.logger.Debug("FavoriteRepository.Add: attempting to add favorite", "user_id", favorite.UserID, "listing_id", favorite.ListingID)

	// Устанавливаем время создания. ID доменной модели будет обновлен после вставки.
	favorite.CreatedAt = time.Now().UTC()

	doc, err := toFavoriteDocument(favorite) // Конвертируем в MongoDB документ
	if err != nil {
		r.logger.Error("FavoriteRepository.Add: failed to convert domain to document", "error", err, "user_id", favorite.UserID, "listing_id", favorite.ListingID)
		return fmt.Errorf("failed to prepare favorite for database: %w", err)
	}
	// doc.ID будет primitive.NilObjectID, если favorite.ID был пуст.

	res, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) { // Требует уникального индекса по user_id, listing_id
			r.logger.Warn("FavoriteRepository.Add: favorite already exists (duplicate key error)", "user_id", favorite.UserID, "listing_id", favorite.ListingID)
			return ErrFavoriteAlreadyExistsDB // Используем ошибку, определенную в этом пакете
		}
		r.logger.Error("FavoriteRepository.Add: InsertOne failed", "error", err, "user_id", favorite.UserID, "listing_id", favorite.ListingID)
		return err
	}

	// Обновляем ID в переданном доменном объекте
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		favorite.ID = oid.Hex()
		r.logger.Info("Favorite added successfully", "id", favorite.ID, "user_id", favorite.UserID, "listing_id", favorite.ListingID)
	} else {
		r.logger.Error("FavoriteRepository.Add: InsertOne returned unexpected ID type", "type", fmt.Sprintf("%T", res.InsertedID))
		return errors.New("failed to retrieve generated favorite ID")
	}
	return nil
}

func (r *FavoriteRepository) Remove(ctx context.Context, userID, listingID string) error {
	r.logger.Debug("FavoriteRepository.Remove: attempting to remove favorite", "user_id", userID, "listing_id", listingID)
	if userID == "" || listingID == "" {
		errMsg := "UserID and ListingID cannot be empty for removing a favorite"
		r.logger.Error("FavoriteRepository.Remove: "+errMsg, "user_id", userID, "listing_id", listingID)
		return errors.New(errMsg)
	}

	filter := bson.M{"user_id": userID, "listing_id": listingID}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		r.logger.Error("FavoriteRepository.Remove: DeleteOne failed", "error", err, "user_id", userID, "listing_id", listingID)
		return err
	}

	if result.DeletedCount == 0 {
		r.logger.Warn("FavoriteRepository.Remove: No favorite found to delete", "user_id", userID, "listing_id", listingID)
		return ErrFavoriteNotFoundDB // Используем ошибку, определенную в этом пакете
	}
	r.logger.Info("Favorite removed successfully", "user_id", userID, "listing_id", listingID)
	return nil
}

func (r *FavoriteRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Favorite, error) {
	r.logger.Debug("FavoriteRepository.FindByUserID: fetching favorites", "user_id", userID)
	if userID == "" {
		errMsg := "UserID cannot be empty for fetching favorites"
		r.logger.Error("FavoriteRepository.FindByUserID: "+errMsg, "user_id", userID)
		return nil, errors.New(errMsg)
	}

	filter := bson.M{"user_id": userID}
	findOptions := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}) // Сортировка: новые сначала

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		r.logger.Error("FavoriteRepository.FindByUserID: Find failed", "error", err, "user_id", userID)
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []*favoriteDocument // Декодируем в слайс документов MongoDB
	if err = cursor.All(ctx, &docs); err != nil {
		r.logger.Error("FavoriteRepository.FindByUserID: Cursor All failed", "error", err, "user_id", userID)
		return nil, err
	}

	r.logger.Info("FavoriteRepository.FindByUserID: Found favorites", "user_id", userID, "count", len(docs))
	return toDomainFavorites(docs), nil // Конвертируем в слайс доменных моделей
}

// FindOneByUserIDAndListingID - полезный метод для проверки существования
func (r *FavoriteRepository) FindOneByUserIDAndListingID(ctx context.Context, userID, listingID string) (*domain.Favorite, error) {
	r.logger.Debug("FavoriteRepository.FindOneByUserIDAndListingID: checking for favorite", "user_id", userID, "listing_id", listingID)
	if userID == "" || listingID == "" {
		return nil, errors.New("UserID and ListingID cannot be empty")
	}
	filter := bson.M{"user_id": userID, "listing_id": listingID}
	var doc favoriteDocument
	err := r.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			r.logger.Debug("FavoriteRepository.FindOneByUserIDAndListingID: favorite not found", "user_id", userID, "listing_id", listingID)
			return nil, ErrFavoriteNotFoundDB // Возвращаем специфичную ошибку
		}
		r.logger.Error("FavoriteRepository.FindOneByUserIDAndListingID: FindOne failed", "error", err, "user_id", userID, "listing_id", listingID)
		return nil, err
	}
	return toDomainFavorite(&doc), nil
}