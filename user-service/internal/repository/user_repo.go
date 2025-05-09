package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Abdurahmanit/GroupProject/user-service/internal/entity"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("email already exists")
	ErrUserNotFound   = errors.New("user not found")
)

type UserRepository struct {
	db    *mongo.Database
	redis *redis.Client
}

func NewUserRepository(db *mongo.Database, redis *redis.Client) *UserRepository {
	// Ensure the email field is indexed as unique in MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"email": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		// Log the error in a production system, but for now, we'll ignore it
		// since it might already exist
	}

	return &UserRepository{
		db:    db,
		redis: redis,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *entity.User) error {
	// Hash the password before storing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	doc := bson.M{
		"_id":               user.ID,
		"username":          user.Username,
		"email":             user.Email,
		"password":          string(hashedPassword),
		"role":              user.Role,
		"is_email_verified": user.IsEmailVerified,
		"is_active":         user.IsActive,
		"created_at":        user.CreatedAt,
		"updated_at":        user.UpdatedAt,
	}

	_, err = r.db.Collection("users").InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrDuplicateEmail
		}
		return err
	}
	return nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	var doc bson.M
	err := r.db.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user := &entity.User{
		ID:              doc["_id"].(string),
		Username:        doc["username"].(string),
		Email:           doc["email"].(string),
		Password:        doc["password"].(string),
		Role:            doc["role"].(string),
		IsEmailVerified: doc["is_email_verified"].(bool),
		IsActive:        doc["is_active"].(bool),
		CreatedAt:       doc["created_at"].(primitive.DateTime).Time(),
		UpdatedAt:       doc["updated_at"].(primitive.DateTime).Time(),
	}
	return user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*entity.User, error) {
	var doc bson.M
	err := r.db.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user := &entity.User{
		ID:              doc["_id"].(string),
		Username:        doc["username"].(string),
		Email:           doc["email"].(string),
		Password:        doc["password"].(string),
		Role:            doc["role"].(string),
		IsEmailVerified: doc["is_email_verified"].(bool),
		IsActive:        doc["is_active"].(bool),
		CreatedAt:       doc["created_at"].(primitive.DateTime).Time(),
		UpdatedAt:       doc["updated_at"].(primitive.DateTime).Time(),
	}
	return user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	update := bson.M{
		"$set": bson.M{
			"username":          user.Username,
			"email":             user.Email,
			"role":              user.Role,
			"is_email_verified": user.IsEmailVerified,
			"is_active":         user.IsActive,
			"updated_at":        user.UpdatedAt,
		},
	}

	result, err := r.db.Collection("users").UpdateOne(ctx, bson.M{"_id": user.ID}, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrDuplicateEmail
		}
		return err
	}
	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"password":   string(hashedPassword),
			"updated_at": time.Now(),
		},
	}

	result, err := r.db.Collection("users").UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, userID string) error {
	// Soft delete by setting is_active to false
	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now(),
		},
	}

	result, err := r.db.Collection("users").UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	// Invalidate any cached token
	return r.InvalidateToken(ctx, userID)
}

func (r *UserRepository) HardDeleteUser(ctx context.Context, userID string) error {
	result, err := r.db.Collection("users").DeleteOne(ctx, bson.M{"_id": userID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrUserNotFound
	}

	// Invalidate any cached token
	return r.InvalidateToken(ctx, userID)
}

func (r *UserRepository) ListUsers(ctx context.Context, skip, limit int64) ([]*entity.User, error) {
	findOptions := options.Find()
	findOptions.SetSkip(skip)
	findOptions.SetLimit(limit)
	findOptions.SetSort(bson.M{"created_at": -1})

	cursor, err := r.db.Collection("users").Find(ctx, bson.M{"is_active": true}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*entity.User
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		user := &entity.User{
			ID:              doc["_id"].(string),
			Username:        doc["username"].(string),
			Email:           doc["email"].(string),
			Password:        doc["password"].(string),
			Role:            doc["role"].(string),
			IsEmailVerified: doc["is_email_verified"].(bool),
			IsActive:        doc["is_active"].(bool),
			CreatedAt:       doc["created_at"].(primitive.DateTime).Time(),
			UpdatedAt:       doc["updated_at"].(primitive.DateTime).Time(),
		}
		users = append(users, user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) SearchUsers(ctx context.Context, query string, skip, limit int64) ([]*entity.User, error) {
	findOptions := options.Find()
	findOptions.SetSkip(skip)
	findOptions.SetLimit(limit)
	findOptions.SetSort(bson.M{"created_at": -1})

	// Search by username or email (case-insensitive)
	filter := bson.M{
		"is_active": true,
		"$or": []bson.M{
			{"username": bson.M{"$regex": query, "$options": "i"}},
			{"email": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	cursor, err := r.db.Collection("users").Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*entity.User
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		user := &entity.User{
			ID:              doc["_id"].(string),
			Username:        doc["username"].(string),
			Email:           doc["email"].(string),
			Password:        doc["password"].(string),
			Role:            doc["role"].(string),
			IsEmailVerified: doc["is_email_verified"].(bool),
			IsActive:        doc["is_active"].(bool),
			CreatedAt:       doc["created_at"].(primitive.DateTime).Time(),
			UpdatedAt:       doc["updated_at"].(primitive.DateTime).Time(),
		}
		users = append(users, user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return users, nil
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
