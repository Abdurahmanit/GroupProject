// internal/adapter/repository/mongodb/models.go
package mongodb

import (
	"fmt"
	"time"

	"github.com/Abdurahmanit/GroupProject/listing-service/internal/listing/domain" // Путь к твоему домену
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// listingDocument - структура для хранения Listing в MongoDB
type listingDocument struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty"` // Используем ObjectID
	UserID      string               `bson:"user_id"`
	CategoryID  string               `bson:"category_id"`
	Title       string               `bson:"title"`
	Description string               `bson:"description"`
	Price       float64              `bson:"price"`
	Status      domain.ListingStatus `bson:"status"`
	Photos      []string             `bson:"photos,omitempty"`
	CreatedAt   time.Time            `bson:"created_at"`
	UpdatedAt   time.Time            `bson:"updated_at"`
}

// favoriteDocument - структура для хранения Favorite в MongoDB
type favoriteDocument struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"` // Используем ObjectID
	UserID    string             `bson:"user_id"`
	ListingID string             `bson:"listing_id"`
	CreatedAt time.Time          `bson:"created_at"`
}

// --- Конвертеры для Listing ---

// toListingDocument конвертирует доменную модель Listing в listingDocument.
// Если domain.Listing.ID пуст, то в listingDocument.ID будет primitive.NilObjectID,
// что заставит MongoDB сгенерировать новый ID при вставке.
// Если domain.Listing.ID не пуст, он конвертируется в ObjectID.
func toListingDocument(l *domain.Listing) (*listingDocument, error) {
	if l == nil {
		return nil, nil
	}

	var docID primitive.ObjectID
	var err error

	if l.ID != "" {
		docID, err = primitive.ObjectIDFromHex(l.ID)
		if err != nil {
			return nil, fmt.Errorf("toListingDocument: invalid ID format '%s' for domain listing: %w", l.ID, err)
		}
	} else {
		// Если ID в домене не задан (например, при создании нового),
		// оставляем docID как primitive.NilObjectID.
		// MongoDB сгенерирует _id при вызове InsertOne с omitempty.
		// Либо можно сгенерировать здесь: docID = primitive.NewObjectID()
		// и тогда нужно обновить l.ID после успешной вставки.
		// Репозиторий будет отвечать за обновление l.ID после InsertOne.
		docID = primitive.NilObjectID // Явное указание, что ID не установлен
	}

	return &listingDocument{
		ID:          docID,
		UserID:      l.UserID,
		CategoryID:  l.CategoryID,
		Title:       l.Title,
		Description: l.Description,
		Price:       l.Price,
		Status:      l.Status,
		Photos:      l.Photos,
		CreatedAt:   l.CreatedAt, // Будет установлено/обновлено в репозитории
		UpdatedAt:   l.UpdatedAt, // Будет установлено/обновлено в репозитории
	}, nil
}

// toDomainListing конвертирует listingDocument из БД в доменную модель Listing.
func toDomainListing(d *listingDocument) *domain.Listing {
	if d == nil {
		return nil
	}
	return &domain.Listing{
		ID:          d.ID.Hex(), // Конвертируем ObjectID в строковое представление
		UserID:      d.UserID,
		CategoryID:  d.CategoryID,
		Title:       d.Title,
		Description: d.Description,
		Price:       d.Price,
		Status:      d.Status,
		Photos:      d.Photos,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

// toDomainListings конвертирует слайс listingDocument в слайс доменных Listing.
func toDomainListings(docs []*listingDocument) []*domain.Listing {
	if docs == nil {
		return nil // Или make([]*domain.Listing, 0) если всегда нужен не-nil слайс
	}
	domainListings := make([]*domain.Listing, 0, len(docs))
	for _, doc := range docs {
		domainListings = append(domainListings, toDomainListing(doc))
	}
	return domainListings
}

// --- Конвертеры для Favorite ---

// toFavoriteDocument конвертирует доменную модель Favorite в favoriteDocument.
func toFavoriteDocument(f *domain.Favorite) (*favoriteDocument, error) {
	if f == nil {
		return nil, nil
	}

	var docID primitive.ObjectID
	var err error

	if f.ID != "" {
		docID, err = primitive.ObjectIDFromHex(f.ID)
		if err != nil {
			return nil, fmt.Errorf("toFavoriteDocument: invalid ID format '%s' for domain favorite: %w", f.ID, err)
		}
	} else {
		docID = primitive.NilObjectID
	}

	return &favoriteDocument{
		ID:        docID,
		UserID:    f.UserID,
		ListingID: f.ListingID,
		CreatedAt: f.CreatedAt, // Будет установлено в репозитории
	}, nil
}

// toDomainFavorite конвертирует favoriteDocument из БД в доменную модель Favorite.
func toDomainFavorite(d *favoriteDocument) *domain.Favorite {
	if d == nil {
		return nil
	}
	return &domain.Favorite{
		ID:        d.ID.Hex(),
		UserID:    d.UserID,
		ListingID: d.ListingID,
		CreatedAt: d.CreatedAt,
	}
}

// toDomainFavorites конвертирует слайс favoriteDocument в слайс доменных Favorite.
func toDomainFavorites(docs []*favoriteDocument) []*domain.Favorite {
	if docs == nil {
		return nil
	}
	domainFavorites := make([]*domain.Favorite, 0, len(docs))
	for _, doc := range docs {
		domainFavorites = append(domainFavorites, toDomainFavorite(doc))
	}
	return domainFavorites
}