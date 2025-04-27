package repository

import (
	"context"
	"time"

	"github.com/Abdurahmanit/GroupProject/user-service/internal/entity"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	db    *mongo.Database
	redis *redis.Client
}

func NewUserRepository(db *mongo.Database, redis *redis.Client) *UserRepository {
	return &UserRepository{
		db:    db,
		redis: redis,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *entity.User) error {
	doc := bson.M{
		"_id":      user.ID,
		"username": user.Username,
		"email":    user.Email,
		"password": user.Password,
	}
	_, err := r.db.Collection("users").InsertOne(ctx, doc)
	return err
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	var doc bson.M
	err := r.db.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&doc)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		ID:       doc["_id"].(string),
		Username: doc["username"].(string),
		Email:    doc["email"].(string),
		Password: doc["password"].(string),
	}
	return user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*entity.User, error) {
	var doc bson.M
	err := r.db.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&doc)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		ID:       doc["_id"].(string),
		Username: doc["username"].(string),
		Email:    doc["email"].(string),
		Password: doc["password"].(string),
	}
	return user, nil
}

func (r *UserRepository) CacheToken(ctx context.Context, userID, token string) error {
	return r.redis.Set(ctx, "token:"+userID, token, 24*time.Hour).Err()
}

func (r *UserRepository) InvalidateToken(ctx context.Context, userID string) error {
	return r.redis.Del(ctx, "token:"+userID).Err()
}

func (r *UserRepository) GetToken(ctx context.Context, userID string) (string, error) {
	token, err := r.redis.Get(ctx, "token:"+userID).Result()
	if err == redis.Nil {
		return "", nil
	}
	return token, err
}
