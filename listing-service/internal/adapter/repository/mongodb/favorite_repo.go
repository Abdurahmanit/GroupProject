package mongodb

import (
	"context"
	"time"

	"github.com/Abdurahmanit/GroupProject/internal/listing/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FavoriteRepository struct {
	collection *mongo.Collection
}

func NewFavoriteRepository(db *mongo.Database) *FavoriteRepository {
	return &FavoriteRepository{collection: db.Collection("favorites")}
}

func (r *FavoriteRepository) Add(ctx context.Context, favorite *domain.Favorite) error {
	favorite.ID = primitive.NewObjectID().Hex()
	favorite.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, favorite)
	return err
}

func (r *FavoriteRepository) Remove(ctx context.Context, userID, listingID string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"user_id": userID, "listing_id": listingID})
	return err
}

func (r *FavoriteRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Favorite, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	var favorites []*domain.Favorite
	err = cursor.All(ctx, &favorites)
	return favorites, err
}