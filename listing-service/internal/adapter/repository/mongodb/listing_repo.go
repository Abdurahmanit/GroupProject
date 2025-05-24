package mongodb

import (
	"context"
	"errors"
	"fmt" // Для форматирования ошибок
	"time"
	"strings"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/listing/domain"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/platform/logger" // Предполагаем, что логгер передается
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ListingRepository struct {
	collection *mongo.Collection
	logger     *logger.Logger // Рекомендуется добавить логгер
}

// NewListingRepository принимает логгер
func NewListingRepository(db *mongo.Database, log *logger.Logger) *ListingRepository {
	return &ListingRepository{
		collection: db.Collection("listings"),
		logger:     log,
	}
}

func (r *ListingRepository) Create(ctx context.Context, listing *domain.Listing) error {
	// Устанавливаем время создания и обновления
	now := time.Now().UTC() // Рекомендуется UTC
	listing.CreatedAt = now
	listing.UpdatedAt = now
	// ID доменной модели будет обновлен после вставки

	doc, err := toListingDocument(listing) // Конвертируем в MongoDB документ
	if err != nil {
		r.logger.Error("Create Listing: failed to convert domain to document", "error", err, "user_id", listing.UserID)
		return fmt.Errorf("failed to prepare listing for database: %w", err)
	}
	// На этом этапе doc.ID может быть primitive.NilObjectID, если listing.ID был пуст.
	// Mongo сгенерирует ID.

	res, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		r.logger.Error("Create Listing: InsertOne failed", "error", err, "user_id", listing.UserID, "title", listing.Title)
		return err
	}

	// Обновляем ID в переданном доменном объекте сгенерированным ID из MongoDB
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		listing.ID = oid.Hex()
		r.logger.Info("Listing created successfully", "id", listing.ID, "user_id", listing.UserID)
	} else {
		r.logger.Error("Create Listing: InsertOne returned unexpected ID type", "type", fmt.Sprintf("%T", res.InsertedID))
		return errors.New("failed to retrieve generated listing ID")
	}
	return nil
}

func (r *ListingRepository) Update(ctx context.Context, listing *domain.Listing) error {
	if listing.ID == "" {
		r.logger.Error("Update Listing: domain listing ID is empty")
		return errors.New("cannot update listing without an ID")
	}

	listing.UpdatedAt = time.Now().UTC() // Обновляем время изменения

	doc, err := toListingDocument(listing) // Конвертируем в MongoDB документ
	if err != nil {
		// doc.ID здесь будет содержать ObjectID, сконвертированный из listing.ID
		r.logger.Error("Update Listing: failed to convert domain to document", "error", err, "listing_id", listing.ID)
		return fmt.Errorf("failed to prepare listing for database update: %w", err)
	}

	filter := bson.M{"_id": doc.ID} // doc.ID уже primitive.ObjectID

	// Создаем bson.M для $set, чтобы обновлять только переданные поля, а не весь документ.
	// Но toListingDocument уже возвращает полный документ. Если мы хотим обновлять только
	// измененные поля, логика должна быть сложнее, или usecase должен передавать только изменения.
	// Пока обновляем весь документ (кроме _id).
	updatePayload := bson.M{
		"user_id":     doc.UserID,
		"category_id": doc.CategoryID,
		"title":       doc.Title,
		"description": doc.Description,
		"price":       doc.Price,
		"status":      doc.Status,
		"photos":      doc.Photos,
		// CreatedAt не обновляем
		"updated_at": doc.UpdatedAt,
	}

	result, err := r.collection.UpdateOne(ctx, filter, bson.M{"$set": updatePayload})
	if err != nil {
		r.logger.Error("Update Listing: UpdateOne failed", "id", listing.ID, "error", err)
		return err
	}

	if result.MatchedCount == 0 {
		r.logger.Warn("Update Listing: No document matched for update", "id", listing.ID)
		return domain.ErrListingNotFound
	}
	if result.ModifiedCount == 0 {
	    r.logger.Info("Update Listing: Document matched but not modified (data might be the same)", "id", listing.ID)
	} else {
	    r.logger.Info("Listing updated successfully", "id", listing.ID)
	}

	return nil
}

func (r *ListingRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		r.logger.Error("Delete Listing: ID is empty")
		return errors.New("cannot delete listing without an ID")
	}
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		r.logger.Error("Delete Listing: Invalid ID format", "id", id, "error", err)
		return fmt.Errorf("invalid ID format for delete '%s': %w", id, err)
	}

	filter := bson.M{"_id": objID}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		r.logger.Error("Delete Listing: DeleteOne failed", "id", id, "error", err)
		return err
	}

	if result.DeletedCount == 0 {
		r.logger.Warn("Delete Listing: No document matched for delete", "id", id)
		return domain.ErrListingNotFound
	}
	r.logger.Info("Listing deleted successfully", "id", id)
	return nil
}

func (r *ListingRepository) FindByID(ctx context.Context, id string) (*domain.Listing, error) {
	if id == "" {
		r.logger.Error("FindByID: ID is empty")
		return nil, errors.New("cannot find listing without an ID")
	}
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		r.logger.Error("FindByID: Invalid ID format", "id", id, "error", err)
		return nil, domain.ErrListingNotFound // Возвращаем доменную ошибку, т.к. такой ID не может существовать
	}

	filter := bson.M{"_id": objID}
	var doc listingDocument
	err = r.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			r.logger.Info("FindByID: Listing not found", "id", id)
			return nil, domain.ErrListingNotFound
		}
		r.logger.Error("FindByID: Error retrieving listing", "id", id, "error", err)
		return nil, err
	}
	r.logger.Debug("FindByID: Listing document found, converting to domain", "id", id)
	return toDomainListing(&doc), nil
}

func (r *ListingRepository) FindByFilter(ctx context.Context, filter domain.Filter) ([]*domain.Listing, int64, error) {
	r.logger.Info("FindByFilter: Searching listings", "filter", fmt.Sprintf("%+v", filter))
	mongoFilter := bson.M{}
	var filterParts []bson.M // Используем $and для надежного комбинирования

	if filter.Query != "" {
		// $text поиск требует текстового индекса. Если его нет, используй $regex.
		// filterParts = append(filterParts, bson.M{"$text": bson.M{"$search": filter.Query}})
		// Альтернатива с $regex для поиска по нескольким полям:
		regexQuery := primitive.Regex{Pattern: filter.Query, Options: "i"}
		filterParts = append(filterParts, bson.M{"$or": []bson.M{
			{"title": regexQuery},
			{"description": regexQuery},
		}})
	}
	if filter.Status != "" {
		filterParts = append(filterParts, bson.M{"status": filter.Status})
	}
	if filter.CategoryID != "" {
		filterParts = append(filterParts, bson.M{"category_id": filter.CategoryID})
	}
	if filter.UserID != "" {
		filterParts = append(filterParts, bson.M{"user_id": filter.UserID})
	}

	priceConditions := bson.M{}
	if filter.MinPrice > 0 {
		priceConditions["$gte"] = filter.MinPrice
	}
	if filter.MaxPrice > 0 {
		priceConditions["$lte"] = filter.MaxPrice
	}
	if len(priceConditions) > 0 {
		filterParts = append(filterParts, bson.M{"price": priceConditions})
	}
	
	if len(filterParts) > 0 {
		mongoFilter["$and"] = filterParts
	}


	findOptions := options.Find()
	if filter.Limit > 0 {
		findOptions.SetLimit(int64(filter.Limit))
		if filter.Page > 0 {
			findOptions.SetSkip(int64(filter.Page-1) * int64(filter.Limit))
		} else {
			findOptions.SetSkip(0)
		}
	}

	if filter.SortBy != "" {
		sortOrderValue := 1 // ASC
		if strings.ToLower(filter.SortOrder) == "desc" {
			sortOrderValue = -1 // DESC
		}
		findOptions.SetSort(bson.D{{Key: filter.SortBy, Value: sortOrderValue}})
	} else {
		findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}}) // Default sort
	}

	cursor, err := r.collection.Find(ctx, mongoFilter, findOptions)
	if err != nil {
		r.logger.Error("FindByFilter: Find failed", "filter", fmt.Sprintf("%+v", filter), "mongo_filter", fmt.Sprintf("%+v", mongoFilter), "error", err)
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var docs []*listingDocument
	if err = cursor.All(ctx, &docs); err != nil {
		r.logger.Error("FindByFilter: Cursor All failed", "error", err)
		return nil, 0, err
	}

	total, err := r.collection.CountDocuments(ctx, mongoFilter)
	if err != nil {
		r.logger.Error("FindByFilter: CountDocuments failed", "mongo_filter", fmt.Sprintf("%+v", mongoFilter), "error", err)
		return nil, 0, err
	}

	r.logger.Info("FindByFilter: Search successful", "found_count", len(docs), "total_count", total)
	return toDomainListings(docs), total, nil
}