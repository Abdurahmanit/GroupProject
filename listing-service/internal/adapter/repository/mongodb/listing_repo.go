package mongodb

import (
	"context"
	"time"

	"github.com/your-org/bike-store/listing-service/internal/listing/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ListingRepository struct {
	collection *mongo.Collection
}

func NewListingRepository(db *mongo.Database) *ListingRepository {
	return &ListingRepository{collection: db.Collection("listings")}
}

func (r *ListingRepository) Create(ctx context.Context, listing *domain.Listing) error {
	listing.ID = primitive.NewObjectID().Hex()
	listing.CreatedAt = time.Now()
	listing.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, listing)
	return err
}

func (r *ListingRepository) Update(ctx context.Context, listing *domain.Listing) error {
	listing.UpdatedAt = time.Now()
	_, err := r.collection.UpdateByID(ctx, listing.ID, bson.M{"$set": listing})
	return err
}

func (r *ListingRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *ListingRepository) FindByID(ctx context.Context, id string) (*domain.Listing, error) {
	var listing domain.Listing
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&listing)
	return &listing, err
}

func (r *ListingRepository) FindByFilter(ctx context.Context, filter domain.Filter) ([]*domain.Listing, error) {
	query := bson.M{}
	if filter.Query != "" {
		query["$text"] = bson.M{"$search": filter.Query}
	}
	if filter.MinPrice > 0 {
		query["price"] = bson.M{"$gte": filter.MinPrice}
	}
	if filter.MaxPrice > 0 {
		if q, ok := query["price"]; ok {
			query["price"] = bson.M{"$gte": q, "$lte": filter.MaxPrice}
		} else {
			query["price"] = bson.M{"$lte": filter.MaxPrice}
		}
	}
	if filter.Status != "" {
		query["status"] = filter.Status
	}

	cursor, err := r.collection.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	var listings []*domain.Listing
	err = cursor.All(ctx, &listings)
	return listings, err
}